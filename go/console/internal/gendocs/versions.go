package gendocs

import (
	"fmt"
	"strings"
)

// DocVersion is the document version shared by every file in docs/api/
// (both locales). It is independent of the CYBERBIZ API version. Bump it on
// every regeneration: patch for description-only fixes, minor for added
// endpoints/fields/events, major for removed or renamed ones.
const DocVersion = "2.0.0"

// releaseNote is one entry of the "## Release Notes" section.
type releaseNote struct {
	Version string
	Date    string
	Changes []string
}

// releaseNotes lists every document version, newest first.
var releaseNotes = []releaseNote{
	{
		Version: "2.0.0",
		Date:    "2026-10-09",
		Changes: []string{
			"Removed the /v1/limit_collections endpoints: the platform does not serve them on app-store-api.cyberbiz.io.",
		},
	},
	{
		Version: "1.1.0",
		Date:    "2026-10-09",
		Changes: []string{
			"Product create and POS batch create document the sku request parameter, which the platform requires for shops with the POS feature.",
		},
	},
	{
		Version: "1.0.1",
		Date:    "2026-10-02",
		Changes: []string{
			"Postman collection descriptions no longer point to repository-internal files.",
		},
	},
	{
		Version: "1.0.0",
		Date:    "2026-09-08",
		Changes: []string{
			"Initial generation from the CYBERBIZ v1 swagger, the v1/v2 Postman collections, the Notion v2 reference and the webhook reference.",
			"Observed corrections applied: nullable fields, money as float, order_number as integer, arrays documented as objects, fields missing from the swagger, Bearer authentication.",
			"Synthetic samples derived from the Golden Files for every operation and event.",
		},
	},
}

// releaseNotesMarkdown renders the release notes section.
func releaseNotesMarkdown(loc *locale) string {
	var b strings.Builder
	b.WriteString("## Release Notes\n")
	for _, n := range releaseNotes {
		fmt.Fprintf(&b, "\n### %s (%s)\n\n", n.Version, n.Date)
		for _, c := range n.Changes {
			fmt.Fprintf(&b, "- %s\n", loc.T(c))
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
