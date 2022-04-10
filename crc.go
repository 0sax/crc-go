package CRC

import (
	"net/http"
)

func NewCRCService(userName, password string) *Service {
	return &Service{
		BaseUrl:  "https://webserver.creditreferencenigeria.net/crcweb/LiveRequestInvoker.asmx/PostRequest",
		UserName: userName,
		Password: password,
		Client: http.Client{},
	}
}

type Service struct {
	BaseUrl string
	UserName string
	Password string
	Client http.Client
}

func (s *Service) SearchByBVN(bvn string, minConfidence int) (cr *CleanedReport, err error)  {
	// initial request
	bdy := RequestObject{
		RequestID:         1,
		RequestParameters: RequestParameterObject{
			ReportParameters: ReportParametersObject{
				ResponseType: 1,
				SubjectType:  1,
				ReportId:     6110,
			},
			InquiryReason:    InquiryReason{
				Code:    1,
			},
			Application:      Application{
				Currency: "NGN",
				Amount:   0,
				Number:   0,
				Product:  017,
			},
		},
		SearchParameters:  &SearchParameterObject{
			SearchType: 4,
			BVN:        bvn,
		},
	}
	var r *ResponseDataPacket
	r, err = s.makeRequest(http.MethodPost,nil,bdy)
	if err != nil {
		return
	}
	
	if r.IsMultiHit() {
		bIDs := r.GetBureauIDsAbove(minConfidence)
		r, err = s.GetMultiHitReport(
			r.GetBureauIDWithHighestConfidenceScore(),
			bIDs, r.RefNo)
		if err != nil {
			return
		}
	}

	cr = r.GetCleanReport(bvn)

	return
}

func (s *Service) GetMultiHitReport(pBureauId string, bureauIds []string, reference string) (response *ResponseDataPacket, err error) {

	bdy := RequestObject{
		RequestID:         1,
		RequestParameters: RequestParameterObject{
			ReportParameters: ReportParametersObject{
				ResponseType: 1,
				SubjectType:  1,
				ReportId:     6110,
			},
			InquiryReason:    InquiryReason{
				Code:    1,
			},
			Application:      Application{
				Currency: "NGN",
				Amount:   0,
				Number:   0,
				Product:  017,
			},
			RequestReference: &RequestReference{ReferenceNo: reference},
		},
	}

	if bureauIds != nil && len(bureauIds) > 1 {
		bdy.RequestParameters.RequestReference.MergeReport =
			&MergeReport{
			PrimaryBureauId: pBureauId,
			BureauId:        bureauIds,
		}
	} else {
		bdy.RequestParameters.RequestReference.BureauID = pBureauId
	}

	return s.makeRequest(http.MethodPost,nil,bdy)
}




/// ------------- /////

type CleanedReport struct {
	BVN string
	NoHit bool
	ID string

	Records []Record

}

type Record struct {
	Institution    string
	Amount         string
	Status         string
	Balance        string
	AmountOverdue  string
	Classification string
	DisbursalDate  string
	MaturityDate   string
	Source         string
	ReportDate     string
	RefreshedOn    string
	BureauIdentifierEntry string
	BureauIdentifierAccount string
}