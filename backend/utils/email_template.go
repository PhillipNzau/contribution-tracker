package utils

import (
	"fmt"
	"time"
)

func BuildOtpEmail(name, otp string) string {
	year := time.Now().Year()
	return fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; background: #f9f9f9; padding: 20px;">
		  <div style="max-width: 500px; margin: auto; background: #ffffff; border-radius: 10px; overflow: hidden; box-shadow: 0 4px 6px rgba(0,0,0,0.1);">
			
			<div style="background: #7378f5; padding: 15px; text-align: center;">
			</div>
			
			<div style="padding: 20px; text-align: center;">
			  <h2 style="color: #333;">Hello %s 👋</h2>
			  <p style="color: #555;">Here’s your one-time password. It is valid for <b>10 minutes</b>.</p>
			  
			  <div style="font-size: 32px; font-weight: bold; color: #7378f5; margin: 20px 0;">
				%s
			  </div>
			  
			  <p style="color: #999;">If you didn’t request this, please ignore this email.</p>
			</div>
			
			<div style="background: #f1f1f1; padding: 15px; text-align: center; font-size: 12px; color: #777;">
			  &copy; %d Vault. All rights reserved.
			</div>
		  </div>
		</div>
	`, name, otp, year)
}
