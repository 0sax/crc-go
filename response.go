package CRC

import (
	"encoding/xml"
	"strconv"
	"time"
)

type ResponseDataPacket struct {
	XMLName   xml.Name `xml:"DATAPACKET"`
	RequestId string   `xml:"REQUEST-ID,attr"`
	RefNo     string   `xml:"REFERENCE-NO,attr"`
	Header    Header   `xml:"HEADER"`
	Body      Body     `xml:"BODY"`
}

func (r *ResponseDataPacket) ResponseCode() string {
	return r.Header.ResponseType.Code
}
func (r *ResponseDataPacket) ResponseDescription() string {
	return r.Header.ResponseType.Description
}
func (r *ResponseDataPacket) ErrorCodes() []string {
	return r.Body.ErrorList.ErrorCodes
}
func (r *ResponseDataPacket) IsError() bool {
	return r.ResponseCode() == "0"
}
func (r *ResponseDataPacket) IsDataPacket() bool {
	return r.ResponseCode() == "1"
}
func (r *ResponseDataPacket) ISNoHit() bool {
	return r.ResponseCode() == "2"
}
func (r *ResponseDataPacket) IsMultiHit() bool {
	return r.ResponseCode() == "3"
}
func (r *ResponseDataPacket) GetBureauIDsAbove(x int) (res []string) {
	for _, y := range r.Body.SearchResultList.SearchResults {
		cs, err := strconv.Atoi(y.ConfidenceScore)
		if err != nil {
			continue
		}
		if cs >= x {
			if res == nil {
				res = []string{}
			}
			res = append(res, y.BureauId)
		}
	}
	return
}
func (r *ResponseDataPacket) GetBureauIDWithHighestConfidenceScore() (h string) {
	hcs := 0
	for _, y := range r.Body.SearchResultList.SearchResults {
		cs, err := strconv.Atoi(y.ConfidenceScore)
		if err != nil {
			continue
		}
		if cs >= hcs {
			h = y.BureauId
			hcs = cs
		}
	}
	return
}
func (r *ResponseDataPacket) GetCleanReport(bvn string) (cr *CleanedReport) {

	var id string
	var records []Record

	if r.IsDataPacket() {
		id = r.Body.ConsumerProfile.ConsumerDetails.Ruid
		records = r.GetCleanRecords()
	}

	return &CleanedReport{
		BVN:     bvn,
		NoHit:   r.ISNoHit(),
		ID:      id,
		Records: records,
	}
}
func (r *ResponseDataPacket) GetCleanRecords() []Record {
	a := r.Body.ConsumerLoans.GetCleanRecords()
	b := r.Body.MortgageLoans.GetCleanRecords()
	c := r.Body.MFBLoans.GetCleanRecords()

	a = append(a, b...)
	a = append(a, c...)

	return a
}

type Header struct {
	ResponseType ResponseType `xml:"RESPONSE-TYPE"`
	//SearchResultList []CRCResultItem `xml:"SEARCH-RESULT-LIST>SEARCH-RESULT-ITEM"`
}
type ResponseType struct {
	Code        string `xml:"CODE,attr"`
	Description string `xml:"DESCRIPTION,attr"`
}
type Body struct {
	ErrorList        ErrorList        `xml:"ERROR-LIST"`
	SearchResultList SearchResultList `xml:"SEARCH-RESULT-LIST"`
	ConsumerProfile  ConsumerProfile  `xml:"CONSUMER_PROFILE"`
	ConsumerLoans    Loans            `xml:"CONSUMER_CREDIT_FACILITY_SI"`
	MFBLoans         Loans            `xml:"MFCONSUMER_CREDIT_FACILITY_SI"`
	MortgageLoans    Loans            `xml:"MGCONSUMER_CREDIT_FACILITY_SI"`
}
type ConsumerProfile struct {
	ConsumerDetails ConsumerDetails `xml:"CONSUMER_DETAILS"`
}
type ConsumerDetails struct {
	Name string `xml:"NAME"`
	Ruid string `xml:"RUID"`
}
type Loans struct {
	Loans []Loan `xml:"CREDIT_DETAILS"`
}

func (l *Loans) GetCleanRecords() (crs []Record) {
	if l == nil || l.Loans == nil || len(l.Loans) == 0 {
		return []Record{}
	}

	for _, loan := range l.Loans {
		if crs == nil {
			crs = []Record{}
		}
		crs = append(crs, loan.GetCleanRecord())
	}
	return
}

type Loan struct {
	ID                  string `xml:"PRIMARY_ROOT_ID"`
	CustomerID          string `xml:"RUID"`
	DateReported        string `xml:"REPORTED_DATE"`
	Lender              string `xml:"PROVIDER_SOURCE"`
	Amount              string `xml:"SANCTIONED_AMOUNT"`
	InstalmentSize      string `xml:"INSTALLMENT_AMOUNT"`
	Balance             string `xml:"BALANCEAMOUNT"`
	CurrentBalance      string `xml:"CURRENT_BALANCE"`
	OverduePrincipal    string `xml:"OVER_DUE_AMT_PRICIPAL"`
	DisbursalDate       string `xml:"FIRST_DISBURSE_DATE"`
	MaturityDate        string `xml:"PLANNED_CLOSURE_DATE"`
	Category            string `xml:"CATEGORY_DESC"`
	AssetClassification string `xml:"ASSET_CLASSIFICATION"`
	Status              string `xml:"ACCOUNT_STATUS"`
	BureauAccountStatus string `xml:"BUREAU_ACC_STATUS"`
	// Asset Classification
	// CG 	& Mortgage 001,002,003,004; standard, substandard, doubtful, lost
	// MFB 1,2,3,4,5; performing, pass and watch, substabdard, doubtful, lost
	// Bureau Account Status
	// 001, 002, 003, 004, 005; open, closed, freezed, inactive, memorandum
}

func (l *Loan) GetCleanRecord() Record {

	return Record{
		Institution:             l.Lender,
		Amount:                  l.Amount,
		Status:                  l.Status,
		Balance:                 l.Balance,
		AmountOverdue:           l.OverduePrincipal,
		Classification:          l.AssetClassification,
		DisbursalDate:           l.DisbursalDate,
		MaturityDate:            l.MaturityDate,
		Source:                  "crc-full",
		ReportDate:              l.DateReported,
		RefreshedOn:             time.Now().Format("02-Jan-2006"),
		BureauIdentifierEntry:   l.ID,
		BureauIdentifierAccount: l.CustomerID,
	}
}

type ErrorList struct {
	ErrorCodes []string `xml:"ERROR-CODE"`
}
type SearchResultList struct {
	SearchResults []SearchResultItem `xml:"SEARCH-RESULT-ITEM"`
}
type SearchResultItem struct {
	BureauId        string `xml:"BUREAU-ID,attr"`
	ConfidenceScore string `xml:"CONFIDENCE-SCORE,attr"`
	Name            string `xml:"NAME,attr"`
}

//type CRCMerge struct {
//	PrimaryBureauID string
//	ReferenceNo     string
//	Results         []CRCResultItem
//	ResponseType    string
//}
