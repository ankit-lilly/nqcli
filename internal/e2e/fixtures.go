package e2e

import (
	"fmt"
	"time"
)

// BuildMinimalPayload creates the smallest valid SdrPayload for integration testing.
func BuildMinimalPayload(trialAlias string) SdrPayload {
	ts := time.Now().UTC().Format(time.RFC3339)
	epochID := "epoch-1"
	activityID := "activity-1"
	visitID := "visit-1"
	instanceID := "instance-1"

	return SdrPayload{
		ID:               fmt.Sprintf("e2e-test-%d", time.Now().UnixMilli()),
		TrialAlias:       trialAlias,
		AmendmentNumber:  "INITIAL",
		TrialTitle:       "E2E Integration Test Trial",
		TrialDescription: "Automated integration test payload",
		TrialPhase:       "PHASE_I_TRIAL",
		StudyType:        "INTERVENTIONAL",
		Compound:         "Test Compound",
		TherapeuticAreas: []string{"Test TA"},
		OpenLabel:        true,
		Indications:      []string{"Test Indication"},
		Collaborators:    []string{"e2e-test@lilly.com"},
		CreatedBy:        "e2e-test@lilly.com",
		UpdatedBy:        "e2e-test@lilly.com",
		SentBy:           "e2e-test@lilly.com",
		System:           "DESIGN_STUDIO",
		SentAt:           ts,
		CreatedAt:        ts,
		UpdatedAt:        ts,
		SoA: []SoAInput{
			{
				Name: "Primary SoA",
				Categories: []CategoryInput{
					{
						ID:           "category-1",
						CategoryName: "General Assessments",
						Activities: []ActivityInput{
							{
								ID:           activityID,
								ActivityName: "Physical Examination",
							},
						},
					},
				},
				Epochs: []EpochInput{
					{
						ID:   epochID,
						Name: "Screening",
						Type: "SCREENING",
					},
				},
				Visits: []VisitInput{
					{
						ID:           visitID,
						VisitLabel:   "Visit 1",
						VisitEpochID: epochID,
						ContactModes: []string{"IN_PERSON"},
					},
				},
				ActivityInstances: []ActivityInstanceInput{
					{
						ID:          instanceID,
						VisitID:     visitID,
						ActivityIDs: []string{activityID},
					},
				},
				Timings: []TimingInput{
					{
						ID:             "timing-1",
						FromInstanceID: instanceID,
						ToInstanceID:   instanceID,
						Type:           "BEFORE",
						Value:          "PT0S",
						RelativeToFrom: "START_TO_START",
					},
				},
				Conditions: []ConditionInput{},
			},
		},
	}
}

// BuildFullPayload creates a comprehensive SdrPayload with multiple epochs,
// visits, activities, and conditions for thorough graph structure testing.
func BuildFullPayload(trialAlias string) SdrPayload {
	ts := time.Now().UTC().Format(time.RFC3339)
	epoch1 := "epoch-full-1"
	epoch2 := "epoch-full-2"
	activity1 := "activity-full-1"
	activity2 := "activity-full-2"
	visit1 := "visit-full-1"
	visit2 := "visit-full-2"
	instance1 := "instance-full-1"
	instance2 := "instance-full-2"

	return SdrPayload{
		ID:               fmt.Sprintf("e2e-full-%d", time.Now().UnixMilli()),
		TrialAlias:       trialAlias,
		AmendmentNumber:  "INITIAL",
		TrialTitle:       "E2E Full Integration Test Trial",
		TrialDescription: "Comprehensive integration test payload",
		TrialPhase:       "PHASE_III_TRIAL",
		StudyType:        "INTERVENTIONAL",
		Compound:         "Test Compound X",
		TherapeuticAreas: []string{"Oncology", "Immunology"},
		OpenLabel:        false,
		Indications:      []string{"Test Condition A", "Test Condition B"},
		Collaborators:    []string{"e2e-test@lilly.com", "e2e-test2@lilly.com"},
		CreatedBy:        "e2e-test@lilly.com",
		UpdatedBy:        "e2e-test@lilly.com",
		SentBy:           "e2e-test@lilly.com",
		System:           "DESIGN_STUDIO",
		SentAt:           ts,
		CreatedAt:        ts,
		UpdatedAt:        ts,
		SoA: []SoAInput{
			{
				Name: "Primary SoA",
				Categories: []CategoryInput{
					{
						ID:           "category-full-1",
						CategoryName: "General Assessments",
						Activities: []ActivityInput{
							{
								ID:           activity1,
								ActivityName: "Physical Examination",
							},
						},
					},
					{
						ID:           "category-full-2",
						CategoryName: "Lab Procedures",
						Activities: []ActivityInput{
							{
								ID:           activity2,
								ActivityName: "Blood Draw",
								Lab: &LabInput{
									LabLocation: "Central",
								},
							},
						},
					},
				},
				Epochs: []EpochInput{
					{
						ID:          epoch1,
						Name:        "Screening",
						Type:        "SCREENING",
						NextEpochID: epoch2,
					},
					{
						ID:              epoch2,
						Name:            "Treatment",
						Type:            "TREATMENT",
						PreviousEpochID: epoch1,
					},
				},
				Visits: []VisitInput{
					{
						ID:           visit1,
						VisitLabel:   "Screening Visit",
						VisitEpochID: epoch1,
						ContactModes: []string{"IN_PERSON"},
					},
					{
						ID:           visit2,
						VisitLabel:   "Treatment Visit 1",
						VisitEpochID: epoch2,
						ContactModes: []string{"IN_PERSON"},
					},
				},
				ActivityInstances: []ActivityInstanceInput{
					{
						ID:          instance1,
						VisitID:     visit1,
						ActivityIDs: []string{activity1},
					},
					{
						ID:          instance2,
						VisitID:     visit2,
						ActivityIDs: []string{activity1, activity2},
					},
				},
				Timings: []TimingInput{
					{
						ID:             "timing-full-1",
						FromInstanceID: instance1,
						ToInstanceID:   instance2,
						Type:           "AFTER",
						Value:          "P2W",
						RelativeToFrom: "START_TO_START",
					},
				},
				Conditions: []ConditionInput{
					{
						ID:           "condition-full-1",
						Text:         "Confirm eligibility before treatment",
						Label:        "Eligibility Check",
						AppliesToIDs: []string{activity1},
					},
				},
			},
		},
	}
}

// BuildInvalidPayload creates a payload with an invalid trialAlias format
// to test synchronous validation rejection.
func BuildInvalidPayload() SdrPayload {
	p := BuildMinimalPayload("INVALID-ALIAS")
	p.TrialAlias = "invalid-alias-format"
	return p
}
