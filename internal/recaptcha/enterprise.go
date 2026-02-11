package recaptcha

import (
	"context"
	"fmt"
	"log"

	recaptcha "cloud.google.com/go/recaptchaenterprise/v2/apiv1"
	recaptchapb "cloud.google.com/go/recaptchaenterprise/v2/apiv1/recaptchaenterprisepb"
)

// VerifyTokenEnterprise verifies a reCAPTCHA token using Google's reCAPTCHA Enterprise API
func VerifyTokenEnterprise(projectID, recaptchaKey, token, recaptchaAction string) (bool, float32, error) {
	// Log verification attempt
	log.Printf("Verifying reCAPTCHA Enterprise token for action: %s", recaptchaAction)

	// Create the reCAPTCHA client
	ctx := context.Background()
	client, err := recaptcha.NewClient(ctx)
	if err != nil {
		log.Printf("Error creating reCAPTCHA client: %v", err)
		return false, 0, fmt.Errorf("error creating reCAPTCHA client: %v", err)
	}
	defer client.Close()

	// Set the properties of the event to be tracked
	event := &recaptchapb.Event{
		Token:   token,
		SiteKey: recaptchaKey,
	}

	assessment := &recaptchapb.Assessment{
		Event: event,
	}

	// Build the assessment request
	request := &recaptchapb.CreateAssessmentRequest{
		Assessment: assessment,
		Parent:     fmt.Sprintf("projects/%s", projectID),
	}

	// Call the API to create an assessment
	response, err := client.CreateAssessment(ctx, request)
	if err != nil {
		log.Printf("Error calling CreateAssessment: %v", err)
		return false, 0, fmt.Errorf("error calling CreateAssessment: %v", err)
	}

	// Check if the token is valid
	if !response.TokenProperties.Valid {
		log.Printf("Invalid token: %v", response.TokenProperties.InvalidReason)
		return false, 0, fmt.Errorf("invalid token: %v", response.TokenProperties.InvalidReason)
	}

	// Check if the expected action was executed
	if response.TokenProperties.Action != recaptchaAction {
		log.Printf("Action mismatch: got %s, expected %s", response.TokenProperties.Action, recaptchaAction)
		return false, 0, fmt.Errorf("action mismatch: got %s, expected %s", response.TokenProperties.Action, recaptchaAction)
	}

	// Get the risk score
	score := response.RiskAnalysis.Score
	log.Printf("reCAPTCHA Enterprise score: %v", score)

	// Log any risk reasons
	if len(response.RiskAnalysis.Reasons) > 0 {
		log.Printf("Risk reasons: %v", response.RiskAnalysis.Reasons)
	}

	// Consider a score of 0.5 or higher as valid (this is customizable)
	isValid := score >= 0.5
	return isValid, score, nil
}
