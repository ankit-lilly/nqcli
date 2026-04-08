package e2e

// SdrPayload mirrors the GraphQL SdrPayload input type from schema/sdr.graphql.
type SdrPayload struct {
	ID               string     `json:"id"`
	TrialAlias       string     `json:"trialAlias"`
	AmendmentNumber  string     `json:"amendmentNumber"`
	TrialTitle       string     `json:"trialTitle"`
	TrialDescription string     `json:"trialDescription,omitempty"`
	TrialPhase       string     `json:"trialPhase"`
	StudyType        string     `json:"studyType"`
	Compound         string     `json:"compound,omitempty"`
	TherapeuticAreas []string   `json:"therapeuticAreas"`
	OpenLabel        bool       `json:"openLabel"`
	Indications      []string   `json:"indications"`
	Collaborators    []string   `json:"collaborators"`
	CreatedBy        string     `json:"createdBy"`
	UpdatedBy        string     `json:"updatedBy"`
	SentBy           string     `json:"sentBy"`
	System           string     `json:"system"`
	SentAt           string     `json:"sentAt"`
	SoA              []SoAInput `json:"SoA"`
	CreatedAt        string     `json:"createdAt"`
	UpdatedAt        string     `json:"updatedAt"`
}

type SoAInput struct {
	Name              string                  `json:"name"`
	Categories        []CategoryInput         `json:"categories"`
	Epochs            []EpochInput            `json:"epochs"`
	Visits            []VisitInput            `json:"visits"`
	ActivityInstances []ActivityInstanceInput `json:"activityInstances"`
	Timings           []TimingInput           `json:"timings"`
	Conditions        []ConditionInput        `json:"conditions"`
}

type CategoryInput struct {
	ID           string          `json:"id"`
	CategoryName string          `json:"categoryName"`
	Activities   []ActivityInput `json:"activities"`
}

type ActivityInput struct {
	ID               string             `json:"id"`
	ActivityName     string             `json:"activityName"`
	ConceptCode      string             `json:"conceptCode,omitempty"`
	StandardCode     *StandardCodeInput `json:"standardCode,omitempty"`
	ActivityComments string             `json:"activityComments,omitempty"`
	ParentActivityID string             `json:"parentActivityId,omitempty"`
	Lab              *LabInput          `json:"lab,omitempty"`
}

type StandardCodeInput struct {
	ID                string                    `json:"id,omitempty"`
	ExtensionAttrs    []ExtensionAttributeInput `json:"extensionAttributes,omitempty"`
	Code              string                    `json:"code,omitempty"`
	CodeSystem        string                    `json:"codeSystem,omitempty"`
	CodeSystemVersion string                    `json:"codeSystemVersion,omitempty"`
	Decode            string                    `json:"decode,omitempty"`
	InstanceType      string                    `json:"instanceType,omitempty"`
}

type ExtensionAttributeInput struct {
	ID           string `json:"id,omitempty"`
	URL          string `json:"url,omitempty"`
	ValueString  string `json:"valueString,omitempty"`
	ValueBoolean *bool  `json:"valueBoolean,omitempty"`
	InstanceType string `json:"instanceType"`
}

type LabInput struct {
	IsBlindToSite    *bool              `json:"isBlindToSite,omitempty"`
	IsBlindToSponsor *bool              `json:"isBlindToSponsor,omitempty"`
	LabLocation      string             `json:"labLocation,omitempty"`
	SpecimenCode     *StandardCodeInput `json:"specimenCode,omitempty"`
}

type EpochInput struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Type            string `json:"type"`
	PreviousEpochID string `json:"previousEpochId,omitempty"`
	NextEpochID     string `json:"nextEpochId,omitempty"`
}

type VisitInput struct {
	ID            string   `json:"id"`
	VisitLabel    string   `json:"visitLabel"`
	VisitEpochID  string   `json:"visitEpochId"`
	VisitComments string   `json:"visitComments,omitempty"`
	ContactModes  []string `json:"contactModes,omitempty"`
}

type ActivityInstanceInput struct {
	ID              string    `json:"id"`
	VisitID         string    `json:"visitId"`
	IsVisitInstance *bool     `json:"isVisitInstance,omitempty"`
	ActivityIDs     []string  `json:"activityIds"`
	Lab             *LabInput `json:"lab,omitempty"`
}

type TimingInput struct {
	ID             string `json:"id"`
	FromInstanceID string `json:"fromInstanceId"`
	ToInstanceID   string `json:"toInstanceId"`
	Type           string `json:"type"`
	Value          string `json:"value"`
	RelativeToFrom string `json:"relativeToFrom"`
	WindowLower    string `json:"windowLower,omitempty"`
	WindowUpper    string `json:"windowUpper,omitempty"`
}

type ConditionInput struct {
	ID           string   `json:"id"`
	Text         string   `json:"text,omitempty"`
	Label        string   `json:"label,omitempty"`
	ContextIDs   []string `json:"contextIds,omitempty"`
	AppliesToIDs []string `json:"appliesToIds,omitempty"`
}

// GraphQL response types.

type SubmitResponse struct {
	Message string `json:"message"`
	IsValid bool   `json:"isValid"`
}

type TrialVersionHistory struct {
	Status              string         `json:"status"`
	TrialAlias          string         `json:"trialAlias"`
	TherapeuticAreas    []string       `json:"therapeuticAreas"`
	Sponsor             string         `json:"sponsor"`
	InitialExportDate   string         `json:"initialExportDate"`
	StudyPhase          string         `json:"studyPhase"`
	StudyType           string         `json:"studyType"`
	DSLatestVersion     string         `json:"dsLatestVersion"`
	SDRLatestVersion    string         `json:"sdrLatestVersion"`
	Author              string         `json:"author"`
	Collaborators       []string       `json:"collaborators"`
	TrialVersionHistory []TrialVersion `json:"trialVersionHistory"`
}

type TrialVersion struct {
	StudyID               string `json:"studyId"`
	TrialAlias            string `json:"trialAlias"`
	DSVersion             string `json:"dsVersion"`
	DSVersionTimestamp    string `json:"dsVersionTimestamp"`
	SDRVersion            string `json:"sdrVersion"`
	SDRIngestionTimestamp string `json:"sdrIngestionTimestamp"`
	Status                string `json:"status"`
	NotificationStatus    bool   `json:"notificationStatus"`
}

type Trial struct {
	TrialAlias            string   `json:"trialAlias"`
	DSVersion             string   `json:"dsVersion"`
	DSVersionTimestamp    string   `json:"dsVersionTimestamp"`
	SDRVersion            string   `json:"sdrVersion"`
	SDRIngestionTimestamp string   `json:"sdrIngestionTimestamp"`
	Status                string   `json:"status"`
	TherapeuticAreas      []string `json:"therapeuticAreas"`
	StudyPhase            string   `json:"studyPhase"`
	StudyType             string   `json:"studyType"`
}

type GraphSummary struct {
	TotalNodes     int          `json:"totalNodes"`
	TotalEdges     int          `json:"totalEdges"`
	TotalNodeTypes int          `json:"totalNodeTypes"`
	Nodes          []NodeDetail `json:"nodes"`
}

type NodeDetail struct {
	Label              string   `json:"label"`
	TotalCount         int      `json:"totalCount"`
	Properties         []string `json:"properties"`
	IncomingEdges      []string `json:"incomingEdges"`
	OutgoingEdges      []string `json:"outgoingEdges"`
	TotalIncomingEdges int      `json:"totalIncomingEdges"`
	TotalOutgoingEdges int      `json:"totalOutgoingEdges"`
}

type DeleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
