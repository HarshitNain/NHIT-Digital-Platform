"""
NHIT Enterprise OTP & Notification Engine
Multi-Channel Dispatcher: Email (Resend, Brevo, Gmail SMTP) & SMS (Twilio, Fast2SMS)
Zero third-party library dependencies (uses Python standard library).
"""

import os
import sys
import json
import ssl
import smtplib
import urllib.request
import urllib.parse
import urllib.error
from email.mime.multipart import MIMEMultipart
from email.mime.text import MIMEText
from typing import Dict, Any, Optional, Tuple

# ═══════════════════════════════════════════════════════════════════════════════
# ENVIRONMENT CONFIGURATION LOADER
# ═══════════════════════════════════════════════════════════════════════════════
def load_env():
    """Load key-value pairs from .env files without external dependencies."""
    paths = [
        os.path.join(os.path.dirname(__file__), ".env"),
        os.path.join(os.path.dirname(os.path.dirname(__file__)), ".env"),
        os.path.join(os.getcwd(), ".env")
    ]
    for p in paths:
        if os.path.exists(p):
            with open(p, "r", encoding="utf-8") as f:
                for line in f:
                    line = line.strip()
                    if line and not line.startswith("#") and "=" in line:
                        k, v = line.split("=", 1)
                        k = k.strip()
                        v = v.strip().strip('"').strip("'")
                        if k and k not in os.environ:
                            os.environ[k] = v

load_env()

# Configuration variables
RESEND_API_KEY = os.environ.get("RESEND_API_KEY", "").strip()
BREVO_API_KEY = os.environ.get("BREVO_API_KEY", "").strip()

# SMTP (Gmail / Custom)
SMTP_HOST = os.environ.get("SMTP_HOST", "smtp.gmail.com").strip()
SMTP_PORT = int(os.environ.get("SMTP_PORT", "465").strip() or 465)
SMTP_USER = os.environ.get("SMTP_USER", "").strip()
SMTP_PASSWORD = os.environ.get("SMTP_PASSWORD", "").strip()
SMTP_FROM_NAME = os.environ.get("SMTP_FROM_NAME", "NHIT Digital Platform Security").strip()

# SMS Gateways
TWILIO_ACCOUNT_SID = os.environ.get("TWILIO_ACCOUNT_SID", "").strip()
TWILIO_AUTH_TOKEN = os.environ.get("TWILIO_AUTH_TOKEN", "").strip()
TWILIO_FROM_NUMBER = os.environ.get("TWILIO_FROM_NUMBER", "").strip()
FAST2SMS_API_KEY = os.environ.get("FAST2SMS_API_KEY", "").strip()

# Sender email defaults
EMAIL_SENDER = os.environ.get("EMAIL_SENDER", "NHIT Security <onboarding@resend.dev>").strip()

# ═══════════════════════════════════════════════════════════════════════════════
# HTML EMAIL TEMPLATE GENERATOR
# ═══════════════════════════════════════════════════════════════════════════════
def build_otp_html_email(user_name: str, otp_code: str, action: str = "Two-Factor Portal Sign-in") -> str:
    """Generate modern, responsive government-grade HTML email."""
    return f"""<!DOCTYPE html>
<html>
<head>
  <meta charset="utf-8">
  <title>NHIT Security Verification</title>
  <style>
    body {{ font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; background-color: #0b1120; margin: 0; padding: 24px; color: #f8fafc; }}
    .container {{ max-width: 560px; margin: 0 auto; background: #0f172a; border: 1px solid #1e293b; border-radius: 16px; overflow: hidden; box-shadow: 0 25px 50px -12px rgba(0, 0, 0, 0.5); }}
    .header {{ background: linear-gradient(135deg, #0a192f 0%, #064e3b 100%); padding: 32px 24px; text-align: center; border-bottom: 1px solid #10b981; }}
    .logo-badge {{ display: inline-block; background: rgba(16, 185, 129, 0.15); border: 1px solid #10b981; color: #34d399; font-size: 13px; font-weight: 700; letter-spacing: 2px; padding: 6px 14px; border-radius: 9999px; text-transform: uppercase; margin-bottom: 12px; }}
    .title {{ margin: 0; color: #ffffff; font-size: 24px; font-weight: 700; letter-spacing: -0.5px; }}
    .subtitle {{ margin: 6px 0 0 0; color: #94a3b8; font-size: 14px; }}
    .body {{ padding: 32px 28px; }}
    .greeting {{ font-size: 16px; color: #e2e8f0; margin-bottom: 16px; }}
    .message {{ font-size: 14px; color: #94a3b8; line-height: 1.6; margin-bottom: 24px; }}
    .otp-box {{ background: #1e293b; border: 2px dashed #10b981; border-radius: 12px; padding: 24px; text-align: center; margin: 24px 0; }}
    .otp-label {{ font-size: 12px; color: #94a3b8; text-transform: uppercase; letter-spacing: 1.5px; font-weight: 600; margin-bottom: 8px; }}
    .otp-code {{ font-size: 38px; font-weight: 800; letter-spacing: 8px; color: #10b981; font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace; }}
    .otp-expiry {{ font-size: 12px; color: #f59e0b; margin-top: 8px; font-weight: 500; }}
    .security-notice {{ background: rgba(245, 158, 11, 0.08); border-left: 4px solid #f59e0b; padding: 12px 16px; border-radius: 4px; font-size: 13px; color: #cbd5e1; margin-top: 24px; line-height: 1.5; }}
    .footer {{ background: #0b1120; padding: 20px 24px; text-align: center; font-size: 12px; color: #64748b; border-top: 1px solid #1e293b; }}
  </style>
</head>
<body>
  <div class="container">
    <div class="header">
      <div class="logo-badge">NHIT Enterprise Security</div>
      <h1 class="title">Security Authorization Code</h1>
      <p class="subtitle">National Highways Infra Trust Digital Operations</p>
    </div>
    <div class="body">
      <div class="greeting">Dear <strong>{user_name}</strong>,</div>
      <div class="message">
        A request has been initiated for <strong>{action}</strong> on the NHIT Digital Platform. Use the one-time security pass-code below to complete your authentication.
      </div>
      <div class="otp-box">
        <div class="otp-label">One-Time Verification Passcode</div>
        <div class="otp-code">{otp_code}</div>
        <div class="otp-expiry">&#9201; Valid for 5 minutes (300 seconds)</div>
      </div>
      <div class="security-notice">
        <strong>Security Warning:</strong> Never share this verification code with anyone, including NHIT IT administrators or project managers. If you did not initiate this sign-in attempt, please notify the Cyber Operations Division immediately.
      </div>
    </div>
    <div class="footer">
      This is an automated system dispatch from NHIT Security Gateway.<br>
      &copy; 2026 National Highways Infra Trust. All rights reserved.
    </div>
  </div>
</body>
</html>"""

# ═══════════════════════════════════════════════════════════════════════════════
# 1. RESEND EMAIL DISPATCHER (Free: 3,000 emails/month, 100/day)
# ═══════════════════════════════════════════════════════════════════════════════
def send_email_via_resend(to_email: str, subject: str, html_content: str) -> Tuple[bool, str]:
    """Send transactional email via Resend REST API."""
    if not RESEND_API_KEY:
        return False, "RESEND_API_KEY not configured"

    url = "https://api.resend.com/emails"
    headers = {
        "Authorization": f"Bearer {RESEND_API_KEY}",
        "Content-Type": "application/json",
        "User-Agent": "NHIT-NotificationService/1.0"
    }
    payload = {
        "from": EMAIL_SENDER,
        "to": [to_email],
        "subject": subject,
        "html": html_content
    }

    try:
        req = urllib.request.Request(url, data=json.dumps(payload).encode("utf-8"), headers=headers, method="POST")
        with urllib.request.urlopen(req, timeout=10) as resp:
            res_data = json.loads(resp.read().decode("utf-8"))
            return True, f"Sent via Resend (ID: {res_data.get('id', 'OK')})"
    except urllib.error.HTTPError as e:
        err_msg = e.read().decode("utf-8")
        return False, f"Resend HTTP {e.code}: {err_msg}"
    except Exception as e:
        return False, f"Resend exception: {str(e)}"

# ═══════════════════════════════════════════════════════════════════════════════
# 2. GMAIL / CUSTOM SMTP DISPATCHER (Free: 500 emails/day via Gmail App Password)
# ═══════════════════════════════════════════════════════════════════════════════
def send_email_via_smtp(to_email: str, subject: str, html_content: str) -> Tuple[bool, str]:
    """Send transactional email via standard SSL/TLS SMTP."""
    if not (SMTP_USER and SMTP_PASSWORD):
        return False, "SMTP_USER or SMTP_PASSWORD not configured"

    msg = MIMEMultipart("alternative")
    msg["Subject"] = subject
    msg["From"] = f"{SMTP_FROM_NAME} <{SMTP_USER}>"
    msg["To"] = to_email

    # Plaintext fallback
    plain_text = f"Your NHIT Security verification code is provided in the HTML version of this message."
    msg.attach(MIMEText(plain_text, "plain"))
    msg.attach(MIMEText(html_content, "html"))

    try:
        if SMTP_PORT == 465:
            context = ssl.create_default_context()
            with smtplib.SMTP_SSL(SMTP_HOST, SMTP_PORT, context=context, timeout=10) as server:
                server.login(SMTP_USER, SMTP_PASSWORD)
                server.sendmail(SMTP_USER, to_email, msg.as_string())
        else:
            with smtplib.SMTP(SMTP_HOST, SMTP_PORT, timeout=10) as server:
                server.starttls(context=ssl.create_default_context())
                server.login(SMTP_USER, SMTP_PASSWORD)
                server.sendmail(SMTP_USER, to_email, msg.as_string())

        return True, f"Sent via SMTP ({SMTP_HOST}:{SMTP_PORT})"
    except Exception as e:
        return False, f"SMTP error: {str(e)}"

# ═══════════════════════════════════════════════════════════════════════════════
# 3. BREVO (SENDINBLUE) DISPATCHER (Free: 300 emails/day, 9,000/month)
# ═══════════════════════════════════════════════════════════════════════════════
def send_email_via_brevo(to_email: str, subject: str, html_content: str) -> Tuple[bool, str]:
    """Send transactional email via Brevo REST API."""
    if not BREVO_API_KEY:
        return False, "BREVO_API_KEY not configured"

    url = "https://api.brevo.com/v3/smtp/email"
    headers = {
        "api-key": BREVO_API_KEY,
        "Content-Type": "application/json",
        "User-Agent": "NHIT-NotificationService/1.0"
    }
    payload = {
        "sender": {"name": "NHIT Security", "email": SMTP_USER or "security@nhit.com"},
        "to": [{"email": to_email}],
        "subject": subject,
        "htmlContent": html_content
    }

    try:
        req = urllib.request.Request(url, data=json.dumps(payload).encode("utf-8"), headers=headers, method="POST")
        with urllib.request.urlopen(req, timeout=10) as resp:
            res_data = json.loads(resp.read().decode("utf-8"))
            return True, f"Sent via Brevo (MessageId: {res_data.get('messageId', 'OK')})"
    except Exception as e:
        return False, f"Brevo error: {str(e)}"

# ═══════════════════════════════════════════════════════════════════════════════
# 4. TWILIO SMS DISPATCHER (Free trial: $15 credit)
# ═══════════════════════════════════════════════════════════════════════════════
def send_sms_via_twilio(phone: str, message: str) -> Tuple[bool, str]:
    """Send SMS via Twilio REST API."""
    if not (TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN and TWILIO_FROM_NUMBER):
        return False, "TWILIO credentials not configured"

    url = f"https://api.twilio.com/2010-04-01/Accounts/{TWILIO_ACCOUNT_SID}/Messages.json"
    auth_str = f"{TWILIO_ACCOUNT_SID}:{TWILIO_AUTH_TOKEN}".encode("utf-8")
    import base64
    b64_auth = base64.b64encode(auth_str).decode("utf-8")

    headers = {
        "Authorization": f"Basic {b64_auth}",
        "Content-Type": "application/x-www-form-urlencoded"
    }
    payload = urllib.parse.urlencode({
        "To": phone,
        "From": TWILIO_FROM_NUMBER,
        "Body": message
    }).encode("utf-8")

    try:
        req = urllib.request.Request(url, data=payload, headers=headers, method="POST")
        with urllib.request.urlopen(req, timeout=10) as resp:
            res_data = json.loads(resp.read().decode("utf-8"))
            return True, f"SMS Sent via Twilio (SID: {res_data.get('sid', 'OK')})"
    except Exception as e:
        return False, f"Twilio SMS error: {str(e)}"

# ═══════════════════════════════════════════════════════════════════════════════
# 5. FAST2SMS DISPATCHER (Free trial for India mobile numbers)
# ═══════════════════════════════════════════════════════════════════════════════
def send_sms_via_fast2sms(phone: str, otp_code: str) -> Tuple[bool, str]:
    """Send OTP SMS via Fast2SMS Indian gateway."""
    if not FAST2SMS_API_KEY:
        return False, "FAST2SMS_API_KEY not configured"

    # Normalize phone: extract last 10 digits
    clean_phone = "".join(filter(str.isdigit, phone))[-10:]
    url = f"https://www.fast2sms.com/dev/bulkV2?authorization={FAST2SMS_API_KEY}&variables_values={otp_code}&route=otp&numbers={clean_phone}"

    try:
        req = urllib.request.Request(url, headers={"cache-control": "no-cache"})
        with urllib.request.urlopen(req, timeout=10) as resp:
            res_data = json.loads(resp.read().decode("utf-8"))
            if res_data.get("return", False):
                return True, "SMS Sent via Fast2SMS"
            return False, f"Fast2SMS error: {res_data.get('message', 'Unknown error')}"
    except Exception as e:
        return False, f"Fast2SMS exception: {str(e)}"

# ═══════════════════════════════════════════════════════════════════════════════
# UNIFIED SMART DISPATCHER
# ═══════════════════════════════════════════════════════════════════════════════
def dispatch_otp(
    recipient: str,
    otp_code: str,
    user_name: str = "Officer",
    action: str = "Portal 2FA Verification",
    phone: Optional[str] = None
) -> Dict[str, Any]:
    """
    Unified entry point for sending OTP across configured channels.
    Tries Resend -> Brevo -> Gmail SMTP in order.
    Optionally dispatches SMS if phone is provided.
    Always falls back gracefully to formatted console output for seamless dev.
    """
    result = {
        "recipient": recipient,
        "otp": otp_code,
        "email_delivered": False,
        "sms_delivered": False,
        "channels": [],
        "details": []
    }

    # 1. Build Email Content
    subject = f"NHIT Security Code: {otp_code} (Valid for 5 minutes)"
    html_content = build_otp_html_email(user_name=user_name, otp_code=otp_code, action=action)

    # 2. Try Email Dispatchers in Priority Order
    email_sent = False

    # A. Resend
    if RESEND_API_KEY:
        ok, msg = send_email_via_resend(recipient, subject, html_content)
        result["details"].append(f"[Resend] {msg}")
        if ok:
            email_sent = True
            result["channels"].append("Email:Resend")

    # B. Brevo
    if not email_sent and BREVO_API_KEY:
        ok, msg = send_email_via_brevo(recipient, subject, html_content)
        result["details"].append(f"[Brevo] {msg}")
        if ok:
            email_sent = True
            result["channels"].append("Email:Brevo")

    # C. Standard SMTP (Gmail)
    if not email_sent and SMTP_USER and SMTP_PASSWORD:
        ok, msg = send_email_via_smtp(recipient, subject, html_content)
        result["details"].append(f"[SMTP] {msg}")
        if ok:
            email_sent = True
            result["channels"].append(f"Email:SMTP({SMTP_HOST})")

    result["email_delivered"] = email_sent

    # 3. SMS Dispatch (If phone provided)
    if phone:
        sms_text = f"Your NHIT Security verification code is: {otp_code}. Valid for 5 minutes. Do not share this code."
        sms_sent = False

        if TWILIO_ACCOUNT_SID and TWILIO_AUTH_TOKEN:
            ok, msg = send_sms_via_twilio(phone, sms_text)
            result["details"].append(f"[Twilio] {msg}")
            if ok:
                sms_sent = True
                result["channels"].append("SMS:Twilio")

        if not sms_sent and FAST2SMS_API_KEY:
            ok, msg = send_sms_via_fast2sms(phone, otp_code)
            result["details"].append(f"[Fast2SMS] {msg}")
            if ok:
                sms_sent = True
                result["channels"].append("SMS:Fast2SMS")

        result["sms_delivered"] = sms_sent

    # 4. Developer Console Sandbox Fallback Log
    tag = "LIVE DISPATCHED" if (result["email_delivered"] or result["sms_delivered"]) else "SANDBOX CONSOLE"
    print("\n" + "=" * 68)
    print(f" [2FA OTP NOTIFIER] Status: {tag}")
    print(f" Recipient : {recipient}" + (f" | Mobile: {phone}" if phone else ""))
    print(f" Passcode  : >>> {otp_code} <<< (TTL: 300s)")
    if result["channels"]:
        print(f" Delivered via: {', '.join(result['channels'])}")
    else:
        print(" Active Mode: Developer Sandbox (No external API keys configured in .env)")
        print(" Tip: Set RESEND_API_KEY or SMTP_USER/SMTP_PASSWORD for live delivery.")
    print("=" * 68 + "\n")

    return result

# ═══════════════════════════════════════════════════════════════════════════════
# CLI SELF-TEST UTILITY
# ═══════════════════════════════════════════════════════════════════════════════
if __name__ == "__main__":
    import argparse
    parser = argparse.ArgumentParser(description="Test NHIT OTP & Notification Engine")
    parser.add_argument("--to", default="admin@nhit.com", help="Recipient email address")
    parser.add_argument("--phone", default=None, help="Recipient phone number (e.g. +919876543210)")
    parser.add_argument("--code", default="739201", help="6-digit test code")
    args = parser.parse_args()

    print("\n[TESTING] Dispatching test verification code...")
    res = dispatch_otp(recipient=args.to, otp_code=args.code, user_name="Test Officer", phone=args.phone)
    print(json.dumps(res, indent=2))
