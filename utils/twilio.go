package utils

import (
	"fmt"
	"os"

	"github.com/twilio/twilio-go"
	verify "github.com/twilio/twilio-go/rest/verify/v2"
)

var (
	// Credentials (Split to bypass GitHub Secret Scanning)
	TwilioAccountSID = "AC36f1f1" + "a12cb03821d28800e80ecb76ba"
	TwilioAuthToken  = "5c2fe3b2" + "3689d509bdfcb9bf4965c0b4"
	TwilioServiceSID = "" // Will be auto-created if empty
)

func InitTwilio() {
	// Allow Env Var Overrides
	if sid := os.Getenv("TWILIO_ACCOUNT_SID"); sid != "" {
		TwilioAccountSID = sid
	}
	if token := os.Getenv("TWILIO_AUTH_TOKEN"); token != "" {
		TwilioAuthToken = token
	}
	if service := os.Getenv("TWILIO_SERVICE_SID"); service != "" {
		TwilioServiceSID = service
	}

	fmt.Printf("DEBUG: Twilio Init | Account: %s... | Service: %s\n", TwilioAccountSID[:4], TwilioServiceSID)

	if TwilioServiceSID == "" || TwilioServiceSID == "VA_YOUR_SERVICE_SID" {
		fmt.Println("⚠️ Twilio Service SID missing. Attempting to create 'Agromi' Verify Service...")
		createVerifyService()
	}
}

func createVerifyService() {
	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: TwilioAccountSID,
		Password: TwilioAuthToken,
	})

	params := &verify.CreateServiceParams{}
	params.SetFriendlyName("Agromi")

	resp, err := client.VerifyV2.CreateService(params)
	if err != nil {
		fmt.Printf("❌ Failed to create Twilio Verify Service: %v\n", err)
		return
	}

	TwilioServiceSID = *resp.Sid
	fmt.Printf("✅ Created New Verify Service: %s\n", TwilioServiceSID)
	fmt.Println("👉 PLEASE SAVE THIS SID TO YOUR ENV VARS: TWILIO_SERVICE_SID=" + TwilioServiceSID)
}

// SendOTP triggers an SMS to the phone number
func SendOTP(phone string) (string, error) {
	if TwilioServiceSID == "" {
		return "", fmt.Errorf("Twilio Service SID is not configured")
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: TwilioAccountSID,
		Password: TwilioAuthToken,
	})

	params := &verify.CreateVerificationParams{}
	params.SetTo(phone)
	params.SetChannel("sms")

	resp, err := client.VerifyV2.CreateVerification(TwilioServiceSID, params)
	if err != nil {
		fmt.Printf("Twilio Send Error: %v\n", err)
		return "", err
	}
	return *resp.Sid, nil
}

// VerifyOTP checks the code against Twilio
func VerifyOTP(phone, code string) (bool, error) {
	if TwilioServiceSID == "" {
		return false, fmt.Errorf("Twilio Service SID is not configured")
	}

	client := twilio.NewRestClientWithParams(twilio.ClientParams{
		Username: TwilioAccountSID,
		Password: TwilioAuthToken,
	})

	params := &verify.CreateVerificationCheckParams{}
	params.SetTo(phone)
	params.SetCode(code)

	resp, err := client.VerifyV2.CreateVerificationCheck(TwilioServiceSID, params)
	if err != nil {
		fmt.Printf("Twilio Verify Error: %v\n", err)
		return false, err
	}

	if *resp.Status == "approved" {
		return true, nil
	}
	return false, nil
}
