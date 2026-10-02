package inbound

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/actinc/cyberbiz-sdk/go/console/internal/db"
	"github.com/actinc/cyberbiz-sdk/go/console/internal/shops"
	"github.com/actinc/cyberbiz-sdk/go/cyberbiz"
	"github.com/actinc/cyberbiz-sdk/go/webhook"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// MaxBody is the webhook body cap.
const MaxBody = 2 << 20

// Receiver handles POST /webhooks/cyberbiz.
type Receiver struct {
	repo  *Repository
	shops *shops.Service
}

// NewReceiver wires the repository and the shop-backed secret resolver.
func NewReceiver(repo *Repository, s *shops.Service) *Receiver {
	return &Receiver{repo: repo, shops: s}
}

// Handle stores a row for every outcome and answers with the status the
// spec assigns to it. The body is never logged at info level.
func (rc *Receiver) Handle(c *gin.Context) {
	ctx := c.Request.Context()
	row := rc.newRow(c)
	body, tooLarge, err := readBody(c.Request)
	if err != nil {
		rc.finish(c, row, http.StatusBadRequest, db.InboundMalformed, "reading body: "+err.Error())
		return
	}
	row.Body = string(body)
	if tooLarge {
		rc.finish(c, row, http.StatusRequestEntityTooLarge, db.InboundMalformed, "body exceeds 2 MiB")
		return
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(body))
	rc.attachShop(ctx, row)

	_, perr := webhook.Parse(ctx, c.Request, rc.shops, webhook.WithMaxBodyBytes(MaxBody))
	status, code := classify(perr)
	row.SignatureValid = status == db.InboundValid || errors.Is(perr, webhook.ErrInvalidDomainSignature)
	row.DomainSignatureValid = domainValidity(row.DomainSignature, perr, row.SignatureValid)
	if status == db.InboundValid && !json.Valid(body) {
		status, code = db.InboundMalformed, http.StatusBadRequest
		perr = errors.New("body is not valid JSON")
	}
	msg := ""
	if perr != nil {
		msg = perr.Error()
	}
	rc.finish(c, row, code, status, msg)
}

func (rc *Receiver) newRow(c *gin.Context) *db.InboundLog {
	h := c.Request.Header
	return &db.InboundLog{
		ShopDomain:      h.Get(webhook.HeaderDomain),
		CustomDomain:    h.Get(webhook.HeaderShopDomain),
		Event:           h.Get(webhook.HeaderEvent),
		Signature:       h.Get(webhook.HeaderSignature),
		DomainSignature: h.Get(webhook.HeaderDomainHMAC),
		Headers:         db.JSONText(headersJSON(cyberbiz.RedactHeaders(h))),
		RemoteAddr:      c.ClientIP(),
	}
}

func (rc *Receiver) attachShop(ctx context.Context, row *db.InboundLog) {
	if row.ShopDomain == "" {
		return
	}
	shop, err := rc.shops.ByDomain(ctx, row.ShopDomain)
	if err != nil {
		log.Error().Err(err).Str("shop_domain", row.ShopDomain).Msg("looking up shop for inbound")
		return
	}
	if shop != nil {
		id := shop.ID
		row.ShopID = &id
	}
}

// finish stores the row (marking duplicates) and writes the response.
func (rc *Receiver) finish(c *gin.Context, row *db.InboundLog, code int, status, reason string) {
	ctx := c.Request.Context()
	row.Status = status
	row.ResponseStatus = code
	if earlier, err := rc.repo.EarliestBySignature(ctx, row.Event, row.Signature); err != nil {
		log.Error().Err(err).Msg("checking duplicate inbound")
	} else if earlier != 0 {
		row.DuplicateOf = &earlier
	}
	if err := rc.repo.Create(ctx, row); err != nil {
		log.Error().Err(err).Msg("storing inbound log")
		c.JSON(http.StatusInternalServerError, gin.H{"ok": false})
		return
	}
	log.Info().Str("event", row.Event).Str("shop_domain", row.ShopDomain).Str("status", status).
		Int("response_status", code).Bool("duplicate", row.DuplicateOf != nil).Str("reason", reason).Msg("inbound")
	if code == http.StatusOK {
		c.JSON(code, gin.H{"ok": true})
		return
	}
	c.JSON(code, gin.H{"ok": false, "status": status})
}

// readBody reads up to MaxBody bytes; tooLarge reports an oversize body.
func readBody(r *http.Request) (body []byte, tooLarge bool, err error) {
	if r.Body == nil {
		return nil, false, nil
	}
	data, err := io.ReadAll(io.LimitReader(r.Body, MaxBody+1))
	if err != nil {
		return nil, false, err
	}
	if len(data) > MaxBody {
		return data[:MaxBody], true, nil
	}
	return data, false, nil
}

func classify(err error) (status string, code int) {
	switch {
	case err == nil:
		return db.InboundValid, http.StatusOK
	case errors.Is(err, webhook.ErrInvalidSignature), errors.Is(err, webhook.ErrInvalidDomainSignature):
		return db.InboundInvalidSignature, http.StatusUnauthorized
	case errors.Is(err, webhook.ErrUnknownShop):
		return db.InboundUnknownShop, http.StatusUnauthorized
	case errors.Is(err, webhook.ErrBodyTooLarge):
		return db.InboundMalformed, http.StatusRequestEntityTooLarge
	default:
		return db.InboundMalformed, http.StatusBadRequest
	}
}

// domainValidity is nil when the header was absent or could not be checked.
func domainValidity(header string, parseErr error, sigValid bool) *bool {
	if header == "" || !sigValid {
		return nil
	}
	v := !errors.Is(parseErr, webhook.ErrInvalidDomainSignature)
	return &v
}

func headersJSON(h http.Header) string {
	b, err := json.Marshal(h)
	if err != nil {
		return "{}"
	}
	return string(b)
}
