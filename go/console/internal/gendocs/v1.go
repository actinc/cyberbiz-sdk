package gendocs

// buildV1 produces docs/api/<locale>/cyberbiz-openapi-v1.yaml from the swagger, corrected by
// what the sweep observed and illustrated with synthetic Samples.
func buildV1(in *inputs, loc *locale, log *report) (*Document, error) {
	doc, err := convertSwagger(in.swagger, log)
	if err != nil {
		return nil, err
	}
	log.countN("strip-html", in.swagger.operationCount())
	log.countN("typo-interger", countSwaggerType(in.swagger, "Interger"))
	log.countN("form-to-json", countFormOperations(in.swagger))
	doc.Info = &Info{
		Title:       "CYBERBIZ API v1",
		Version:     DocVersion,
		Description: infoDescription(v1Intro, loc),
	}
	applyCorrections(doc, in.observed, log)
	addSharedComponents(doc, log)
	applySharedParameters(doc, log)
	attachExamples(doc, in.golden, in.postmanV1, loc, log)
	applyErrorResponses(doc, in.golden, log)
	localizeDocument(doc, loc, log)
	return doc, nil
}

const v1Intro = `# CYBERBIZ API v1

Every ` + "`/v1`" + ` operation of the CYBERBIZ e-commerce platform, generated from the official
Swagger 2.0 definition and corrected against recorded live responses (Golden Files).
Where the two disagree the recorded behaviour wins; see the ` + "`x-cyberbiz-observed`" + ` marker on
fields the swagger does not document.`

func countSwaggerType(sw *swaggerDoc, typ string) int {
	n := 0
	var walk func(s *swSchema)
	walk = func(s *swSchema) {
		if s == nil {
			return
		}
		if s.Type == typ {
			n++
		}
		for _, k := range s.Properties.Keys() {
			walk(s.Properties.Get(k))
		}
		walk(s.Items)
	}
	for _, d := range sw.Definitions {
		walk(d)
	}
	return n
}

func countFormOperations(sw *swaggerDoc) int {
	n := 0
	for _, ops := range sw.Paths {
		for _, op := range ops {
			for _, p := range op.Parameters {
				if p.In == "formData" {
					n++
					break
				}
			}
		}
	}
	return n
}
