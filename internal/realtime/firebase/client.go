package firebase

import (
	"context"
	"errors"
	"log"

	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"google.golang.org/api/option"
)

var FcmClient *messaging.Client

func InitFirebase() {
	// Path updated to point to the file in the same directory (or adjust relative path if needed based on execution context)
	// Since the app runs from root, the path should be relative to root or absolute.
	// We will move the json file to internal/realtime/firebase/
	// so the path from root is internal/realtime/firebase/teka-pro-firebase-adminsdk-fbsvc-c424882af5.json
	opt := option.WithCredentialsFile("internal/realtime/firebase/teka-pro-firebase-adminsdk-fbsvc-1c6ec4cb07.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing firebase app: %v", err)
	}

	client, err := app.Messaging(context.Background())
	if err != nil {
		log.Fatalf("error getting FCM client: %v", err)
	}

	FcmClient = client
	log.Println("✅ Firebase initialized")
}

func SendFCM(token, title, body string) {
	if FcmClient == nil {
		log.Println("FCM not initialized")
		return
	}
	msg := &messaging.Message{
		Token: token,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			Notification: &messaging.AndroidNotification{
				Sound: "default",
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound:            "default",
					ContentAvailable: true,
					MutableContent:   true,
				},
			},
			Headers: map[string]string{
				"apns-priority": "10",
			},
		},
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
	}
	_, err := FcmClient.Send(context.Background(), msg)
	if err != nil {
		log.Println("❌ FCM send error:", err)
	}
}

// -------------------------
// REALTIME: FCM UTILS
// -------------------------
func SendFCMToTokens(
	ctx context.Context,
	tokens []string,
	title string,
	body string,
	data map[string]string,
) map[string]error {

	results := make(map[string]error)

	if FcmClient == nil {
		log.Println("FCM not initialized")
		for _, t := range tokens {
			results[t] = errors.New("FCM not initialized")
		}
		return results
	}

	for _, token := range tokens {
		msg := &messaging.Message{
			Token: token,
			Android: &messaging.AndroidConfig{
				Priority: "high",
				Notification: &messaging.AndroidNotification{
					Sound:     "default",
					ChannelID: "high_priority_channel", // Ensure this matches client implementation
				},
			},
			APNS: &messaging.APNSConfig{
				Payload: &messaging.APNSPayload{
					Aps: &messaging.Aps{
						Alert: &messaging.ApsAlert{
							Title: title,
							Body:  body,
						},
						Sound:            "default",
						ContentAvailable: true,
						MutableContent:   true,
					},
				},
				Headers: map[string]string{
					"apns-priority": "10", // 10 is high priority
				},
			},
			Notification: &messaging.Notification{
				Title: title,
				Body:  body,
			},
			Data: data,
		}

		res, err := FcmClient.Send(ctx, msg)
		if err != nil {
			log.Println("❌ FCM failed:", err, "token:", token)
			results[token] = err

			// 🔥 Auto-Cleanup: Hapus token jika sudah tidak valid (404 / Not Found)
			if messaging.IsRegistrationTokenNotRegistered(err) {
				log.Printf("🧹 Cleaning up invalid token: %s", token)
				_ = DeleteFCMToken(token)
			}
		} else {
			log.Println("✅ FCM sent, messageID:", res, "token:", token)
			results[token] = nil
		}
	}

	return results
}
