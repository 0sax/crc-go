package CRC

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

type RequestObject struct {
	XMLName           xml.Name               `xml:"REQUEST"`
	RequestID         int                    `xml:"REQUEST_ID,attr"` // if this duesnt work, try changing the type to xml.Attr
	RequestParameters RequestParameterObject `xml:"REQUEST_PARAMETERS"`
	SearchParameters  *SearchParameterObject  `xml:"SEARCH_PARAMETERS,omitempty"`
}
type RequestParameterObject struct {
	XMLName          xml.Name               `xml:"REQUEST_PARAMETERS"`
	ReportParameters ReportParametersObject `xml:"REPORT_PARAMETERS"`
	InquiryReason    InquiryReason          `xml:"INQUIRY_REASON"`
	Application      Application            `xml:"APPLICATION"`
	RequestReference *RequestReference `xml:"REQUEST_REFERENCE,omitempty"`
}
type RequestReference struct {
	XMLName          xml.Name               `xml:"REQUEST_REFERENCE"`
	ReferenceNo string `xml:"REFERENCE-NO,attr"`
	BureauID string `xml:"BUREAU-ID,attr,omitempty"`
	MergeReport *MergeReport `xml:"MERGE_REPORT,omitempty"`
}
type MergeReport struct {
	XMLName          xml.Name               `xml:"MERGE_REPORT"`
	PrimaryBureauId string `xml:"PRIMARY-BUREAU-ID,attr"`
	BureauId []string `xml:"BUREAU_ID"`
}
type InquiryReason struct {
	XMLName xml.Name `xml:"INQUIRY_REASON"`
	Code    int      `xml:"CODE,attr"` // if this duesnt work, try changing the type to xml.Attr
}
type Application struct {
	Currency string `xml:"CURRENCY,attr"` // if this duesnt work, try changing the type to xml.Attr
	Amount   int    `xml:"AMOUNT,attr"`   // if this duesnt work, try changing the type to xml.Attr
	Number   int    `xml:"NUMBER,attr"`   // if this duesnt work, try changing the type to xml.Attr
	Product  int    `xml:"PRODUCT,attr"`  // if this duesnt work, try changing the type to xml.Attr
}
type ReportParametersObject struct {
	XMLName      xml.Name `xml:"REPORT_PARAMETERS"`
	ResponseType int      `xml:"RESPONSE_TYPE,attr"` // if this duesnt work, try changing the type to xml.Attr
	SubjectType  int      `xml:"SUBJECT_TYPE,attr"`  // if this duesnt work, try changing the type to xml.Attr
	ReportId     int      `xml:"REPORT_ID,attr"`     // if this duesnt work, try changing the type to xml.Attr
}
type SearchParameterObject struct {
	XMLName    xml.Name `xml:"SEARCH_PARAMETERS"`
	SearchType int      `xml:"SEARCH-TYPE,attr"`
	BVN        string   `xml:"BVN_NO"`
}
type reqWrapper struct {
	Username string
	Password string
	Request  string
}

func (r *reqWrapper) encode() string {
	data := url.Values{}
	data.Set("strUserID", r.Username)
	data.Set("strPassword", r.Password)
	data.Set("strRequest", r.Request)
	return data.Encode()
}
func (s *Service) makeRequest(method string, headers map[string]interface{}, body interface{}) (rdp *ResponseDataPacket, err error) {

	//convert request to XML
	b, err := xml.Marshal(body)
	if err != nil {
		fmt.Printf("error at point 2.2: %v\n", err) //debug delete
		return
	}
	rw := &reqWrapper{
		Username: s.UserName,
		Password: s.Password,
		Request:  string(b),
	}

	encReq := rw.encode()
	req, err := http.NewRequest(method, s.BaseUrl, strings.NewReader(encReq))
	if err != nil {
		fmt.Printf("error at point 2: %v\n", err) //debug delete
		return
	}

	// Add headers to request
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(encReq)))
	for k, v := range headers {
		req.Header.Set(k, v.(string))
	}


	resp, err := s.Client.Do(req)
	if err != nil {
		fmt.Printf("error at point 3: %v\n", err) //debug delete
		return
	}
	defer resp.Body.Close()

	bdy, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("error at point 4: %v\n", err) //debug delete
		return
	}

	//fmt.Printf("dabahdy: %v\n\n", string(bdy)) //debug delete

	var rawString []byte
	err = xml.Unmarshal(bdy, &rawString)
	if err != nil {
		fmt.Println("C..") //DEBUG delete
		return
	}
	err = xml.Unmarshal(rawString, &rdp)
	if err != nil {
		fmt.Printf("error at point 6a: %v\n", err) //debug delete
		return
	}

	if rdp.IsError() {
		err = fmt.Errorf("error description:'%v'\nerror code(s):'%v'",
			rdp.ResponseDescription(),strings.Join(rdp.ErrorCodes(),", "))
	}

	return

}
