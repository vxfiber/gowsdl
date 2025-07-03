package gowsdl

var serverTmpl = `

var ErrWSDLUndefined = errors.New("server was unable to process request. --> Object reference not set to an instance of an object")

type SOAPEnvelopeRequest struct {
	XMLName xml.Name ` + "`" + `xml:"http://schemas.xmlsoap.org/soap/envelope/ Envelope"` + "`" + `
	Body SOAPBodyRequest
}

type SOAPBodyRequest struct {
	XMLName xml.Name ` + "`" + `xml:"http://schemas.xmlsoap.org/soap/envelope/ Body"` + "`" + `
	{{range .}}
		{{range .Operations}}
				{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}} ` + `
  				{{$requestType}} *{{$requestType}} ` + "`" + `xml:",omitempty"` + "`" + `
		{{end}}
	{{end}}
}

type SOAPEnvelopeResponse struct { ` + `
	XMLName    xml.Name` + "`" + `xml:"soap:Envelope"` + "`" + `
	PrefixSoap string  ` + "`" + `xml:"xmlns:soap,attr"` + "`" + `
	PrefixXsi  string  ` + "`" + `xml:"xmlns:xsi,attr"` + "`" + `
	PrefixXsd  string  ` + "`" + `xml:"xmlns:xsd,attr"` + "`" + `

	Body SOAPBodyResponse
}

func NewSOAPEnvelopResponse() *SOAPEnvelopeResponse {
	return &SOAPEnvelopeResponse{
		PrefixSoap: "http://schemas.xmlsoap.org/soap/envelope/",
		PrefixXsd:  "http://www.w3.org/2001/XMLSchema",
		PrefixXsi:  "http://www.w3.org/2001/XMLSchema-instance",
	}
}

type Fault struct { ` + `
	XMLName xml.Name ` + "`" + `xml:"SOAP-ENV:Fault"` + "`" + `
	Space   string   ` + "`" + `xml:"xmlns:SOAP-ENV,omitempty,attr"` + "`" + `

	Code   string    ` + "`" + `xml:"faultcode,omitempty"` + "`" + `
	String string    ` + "`" + `xml:"faultstring,omitempty"` + "`" + `
	Actor  string 	 ` + "`" + `xml:"faultactor,omitempty"` + "`" + `
	Detail string    ` + "`" + `xml:"detail,omitempty"` + "`" + `
}


type SOAPBodyResponse struct { ` + `
	XMLName xml.Name   ` + "`" + `xml:"soap:Body"` + "`" + `
	Fault   *Fault ` + "`" + `xml:",omitempty"` + "`" + `
{{range .}}
	{{range .Operations}}
		{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
		{{$responseType}} *{{$responseType}} ` + "`" + `xml:",omitempty"` + "`" + `
	{{end}}
{{end}}

}

type SOAPService interface {
	{{range .}}
		{{range .Operations}}
			{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
			{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
			{{$requestTypeSource := findType .Input.Message | replaceReservedWords }}
			{{$requestType}}Func(request *{{$requestType}}) (*{{$responseType}}, error)
		{{end}}
	{{end}}
}

func NewSOAPEndpoint(svc SOAPService, useWSSecurity bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := SOAPEnvelopeRequest{}
		w.Header().Add("Content-Type", "text/xml; charset=utf-8")

		if r.Method == http.MethodGet {
			w.Write([]byte(wsdl))
			return
		}

		resp := NewSOAPEnvelopResponse()
		defer func() {
			if r := recover(); r != nil {
				resp.Body.Fault = &Fault{}
				resp.Body.Fault.Space = "http://schemas.xmlsoap.org/soap/envelope/"
				resp.Body.Fault.Code = "soap:Server"
				resp.Body.Fault.Detail = fmt.Sprintf("%v", r)
				resp.Body.Fault.String = fmt.Sprintf("%v", r)
			}
			xml.NewEncoder(w).Encode(resp)
		}()

		header := r.Header.Get("Content-Type")
		if strings.Contains(header, "application/soap+xml") {
			panic("Could not find an appropriate Transport Binding to invoke.")
		}

		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}

		if useWSSecurity {
			doc := etree.NewDocument()
			if err := doc.ReadFromBytes(rawBody); err != nil {
				panic(err)
			}
			
			if err := validate(doc); err != nil {
				panic(err)
			}
		}

		if err := xml.Unmarshal(rawBody, &request); err != nil {
			panic(err)
		}
		
		{{range .}}
			{{range .Operations}}
				{{$responseType := findType .Output.Message | replaceReservedWords | makePublic}}
				{{$requestType := findType .Input.Message | replaceReservedWords | makePublic}}
				{{$requestTypeSource := findType .Input.Message | replaceReservedWords }}
				if request.Body.{{$requestType}} != nil {
					resp.Body.{{$responseType}}, err = svc.{{$requestType}}Func(request.Body.{{$requestType}})
					if err != nil {
						panic(err)
					}
					return
				}
			{{end}}
		{{end}}

		panic(ErrWSDLUndefined)
	}
}

func validate(doc *etree.Document) error {
	ctx := dsig.NewDefaultValidationContext(&soap.LocalFileStore{})

	// It is important to only use the returned validated element. // Leos note: NO
	// See: https://www.w3.org/TR/xmldsig-bestpractices/#check-what-is-signed
	_, err := ctx.ValidateBody(doc.Root())
	if err != nil {
		return err
	}

	return nil
}
`
