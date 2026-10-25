import React, { useState, useEffect, useRef } from 'react';
import {
  Shield,
  Lock,
  Mail,
  ArrowRight,
  CheckCircle2,
  User,
  KeyRound,
  ArrowLeft,
  Clock,
  RefreshCw,
  Smartphone,
} from 'lucide-react';
import { api, setAuthToken } from '../services/api';

interface LoginViewProps {
  onLoginSuccess: (user: {
    user_id: string;
    email: string;
    name: string;
    roles: string[];
    permissions: string[];
    auth_method?: 'password' | 'password_otp' | 'google_sso';
    portalMode: 'admin' | 'user';
  }) => void;
  onBackToProduct?: () => void;
  initialPortalMode?: 'admin' | 'user';
}

export const LoginView: React.FC<LoginViewProps> = ({
  onLoginSuccess,
  onBackToProduct,
  initialPortalMode = 'admin',
}) => {
  const [portalMode, setPortalMode] = useState<'admin' | 'user'>(initialPortalMode);
  const [authTab, setAuthTab] = useState<'password' | 'direct_otp'>('password');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [enable2FA, setEnable2FA] = useState(true);

  // 2FA OTP Challenge States & Password Change
  const [loginStep, setLoginStep] = useState<'credentials' | 'otp_verify' | 'force_change_password'>('credentials');
  const [sessionTempToken, setSessionTempToken] = useState<string>('');
  const [devOtpPreview, setDevOtpPreview] = useState<string>('');
  const [otpDigits, setOtpDigits] = useState<string[]>(['', '', '', '', '', '']);
  const [otpCountdown, setOtpCountdown] = useState<number>(300);

  // Mandatory Password Change (OPEN-03)
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const [loading, setLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState('');
  const [infoMsg, setInfoMsg] = useState('');

  const digitInputRefs = useRef<(HTMLInputElement | null)[]>([]);

  useEffect(() => {
    setPortalMode(initialPortalMode);
  }, [initialPortalMode]);

  // OTP Countdown Timer
  useEffect(() => {
    if (loginStep !== 'otp_verify' || otpCountdown <= 0) return;
    const interval = setInterval(() => {
      setOtpCountdown((prev) => (prev > 0 ? prev - 1 : 0));
    }, 1000);
    return () => clearInterval(interval);
  }, [loginStep, otpCountdown]);


  const formatTimer = (sec: number) => {
    const m = Math.floor(sec / 60);
    const s = sec % 60;
    return `${m.toString().padStart(2, '0')}:${s.toString().padStart(2, '0')}`;
  };

  // Step 1: Submit Credentials / Request OTP
  const handleLoginSubmit = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    setErrorMsg('');
    setInfoMsg('');
    setLoading(true);

    try {
      if (authTab === 'password') {
        const res = await api.login({
          login: email,
          password,
          enable_2fa: enable2FA,
        });

        if (res.status === 'PASSWORD_CHANGE_REQUIRED' && res.session_temp_token) {
          setSessionTempToken(res.session_temp_token);
          setLoginStep('force_change_password');
          setInfoMsg(res.message || 'Mandatory password update required before accessing the platform.');
        } else if (res.status === '2FA_REQUIRED' && res.session_temp_token) {
          setSessionTempToken(res.session_temp_token);
          setDevOtpPreview(res.dev_otp_preview || '');
          setOtpDigits(['', '', '', '', '', '']);
          setOtpCountdown(res.expires_in || 300);
          setLoginStep('otp_verify');
          setInfoMsg(`Two-Factor security code generated for ${email}`);
          setTimeout(() => digitInputRefs.current[0]?.focus(), 100);
        } else if (res.token) {
          // Direct login (when 2FA disabled)
          setAuthToken(res.token);
          onLoginSuccess({
            user_id: res.user_id!,
            email: res.email!,
            name: res.name!,
            roles: res.roles || ['employee'],
            permissions: res.permissions || [],
            auth_method: 'password',
            portalMode,
          });
        }
      } else {
        // Passwordless direct OTP request
        const res = await api.requestLoginOTP(email);
        if (res.status === '2FA_REQUIRED' && res.session_temp_token) {
          setSessionTempToken(res.session_temp_token);
          setDevOtpPreview(res.dev_otp_preview || '');
          setOtpDigits(['', '', '', '', '', '']);
          setOtpCountdown(res.expires_in || 300);
          setLoginStep('otp_verify');
          setInfoMsg(`Login OTP code sent to ${email}`);
          setTimeout(() => digitInputRefs.current[0]?.focus(), 100);
        }
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Authentication failed. Please verify your credentials.');
    } finally {
      setLoading(false);
    }
  };

  // Step 2: Verify 6-digit OTP
  const handleVerifyOTP = async (e?: React.FormEvent) => {
    if (e) e.preventDefault();
    const code = otpDigits.join('');
    if (code.length < 6) {
      setErrorMsg('Please enter all 6 digits of the OTP code.');
      return;
    }

    setErrorMsg('');
    setLoading(true);

    try {
      const res = await api.verifyLoginOTP(sessionTempToken, code);
      if (res.status === 'PASSWORD_CHANGE_REQUIRED' && res.session_temp_token) {
        setSessionTempToken(res.session_temp_token);
        setLoginStep('force_change_password');
        setInfoMsg(res.message || 'Mandatory password update required before accessing the platform.');
      } else if (res.token) {
        setAuthToken(res.token);
        onLoginSuccess({
          user_id: res.user_id!,
          email: res.email!,
          name: res.name!,
          roles: res.roles || ['employee'],
          permissions: res.permissions || [],
          auth_method: 'password_otp',
          portalMode,
        });
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'OTP verification failed.');
    } finally {
      setLoading(false);
    }
  };

  // Step 3: Handle Mandatory Password Change (OPEN-03)
  const handleChangePassword = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!newPassword || newPassword.length < 8) {
      setErrorMsg('New password must be at least 8 characters long.');
      return;
    }
    if (newPassword !== confirmPassword) {
      setErrorMsg('New password and confirmation password do not match.');
      return;
    }

    setErrorMsg('');
    setLoading(true);

    try {
      const res = await api.changePassword({
        new_password: newPassword,
        confirm_password: confirmPassword,
        session_temp_token: sessionTempToken,
      });

      if (res.token) {
        setAuthToken(res.token);
        onLoginSuccess({
          user_id: res.user_id!,
          email: res.email!,
          name: res.name!,
          roles: res.roles || ['employee'],
          permissions: res.permissions || [],
          auth_method: 'password',
          portalMode,
        });
      } else {
        setLoginStep('credentials');
        setInfoMsg('Password updated successfully. Please sign in with your new password.');
        setPassword(newPassword);
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Password update failed.');
    } finally {
      setLoading(false);
    }
  };

  // Google SSO Trigger with Google Identity Services GIS Prompt (OPEN-11)
  const handleGoogleSSO = async () => {
    setErrorMsg('');
    const gWindow = window as any;
    if (gWindow.google?.accounts?.id) {
      try {
        gWindow.google.accounts.id.initialize({
          client_id: '984321045981-nhit-digital-platform.apps.googleusercontent.com',
          callback: async (credentialResponse: any) => {
            if (credentialResponse?.credential) {
              setLoading(true);
              try {
                const res = await api.googleSSO({ credential: credentialResponse.credential });
                if (res.token) {
                  setAuthToken(res.token);
                  onLoginSuccess({
                    user_id: res.user_id!,
                    email: res.email!,
                    name: res.name!,
                    roles: res.roles || ['employee'],
                    permissions: res.permissions || [],
                    auth_method: 'google_sso',
                    portalMode,
                  });
                }
              } catch (err: any) {
                setErrorMsg(err.message || 'Google SSO verification failed.');
              } finally {
                setLoading(false);
              }
            }
          },
        });
        gWindow.google.accounts.id.prompt((notification: any) => {
          if (notification && (notification.isNotDisplayed?.() || notification.isSkippedMoment?.())) {
            triggerDirectGoogleSSO();
          }
        });
        return;
      } catch {
        // Fallback to direct request
      }
    }
    await triggerDirectGoogleSSO();
  };

  const triggerDirectGoogleSSO = async () => {
    setLoading(true);
    try {
      const res = await api.googleSSO({
        email,
        name: email === 'admin@nhit.com' ? 'NHIT Super Administrator' : 'Google Officer',
      });
      if (res.token) {
        setAuthToken(res.token);
        onLoginSuccess({
          user_id: res.user_id!,
          email: res.email!,
          name: res.name!,
          roles: res.roles || ['employee'],
          permissions: res.permissions || [],
          auth_method: 'google_sso',
          portalMode,
        });
      }
    } catch (err: any) {
      if (err.message && err.message.toLowerCase().includes('credential')) {
        setErrorMsg('Google Single Sign-On requires verified OAuth credentials. In production environments, please log in with your assigned NHIT credentials below.');
      } else {
        setErrorMsg(err.message || 'Google SSO failed.');
      }
    } finally {
      setLoading(false);
    }
  };

  const handleQuickDemoLogin = async () => {
    setEmail('admin@nhit.com');
    setPassword('password123');
    setEnable2FA(false);
    setLoading(true);
    setErrorMsg('');
    try {
      const res = await api.login({
        login: 'admin@nhit.com',
        password: 'password123',
        enable_2fa: false,
      });
      if (res.token) {
        setAuthToken(res.token);
        onLoginSuccess({
          user_id: res.user_id!,
          email: res.email!,
          name: res.name || 'NHIT Super Administrator',
          roles: res.roles || ['super_admin'],
          permissions: res.permissions || ['*'],
          auth_method: 'password',
          portalMode,
        });
      }
    } catch (err: any) {
      setErrorMsg(err.message || 'Demo login failed');
    } finally {
      setLoading(false);
    }
  };

  // OTP Digit Input Handlers
  const handleDigitChange = (index: number, val: string) => {
    const cleanVal = val.replace(/[^0-9]/g, '');
    if (!cleanVal) {
      const newDigits = [...otpDigits];
      newDigits[index] = '';
      setOtpDigits(newDigits);
      return;
    }

    // Single digit entry
    const char = cleanVal.slice(-1);
    const newDigits = [...otpDigits];
    newDigits[index] = char;
    setOtpDigits(newDigits);

    // Auto focus next box
    if (index < 5) {
      digitInputRefs.current[index + 1]?.focus();
    }
  };

  const handleDigitKeyDown = (index: number, e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Backspace' && !otpDigits[index] && index > 0) {
      digitInputRefs.current[index - 1]?.focus();
    }
  };

  const handleDigitPaste = (e: React.ClipboardEvent<HTMLInputElement>) => {
    e.preventDefault();
    const pasted = e.clipboardData.getData('text').replace(/[^0-9]/g, '').slice(0, 6);
    if (!pasted) return;
    const newDigits = [...otpDigits];
    for (let i = 0; i < pasted.length; i++) {
      newDigits[i] = pasted[i];
    }
    setOtpDigits(newDigits);
    const nextIdx = Math.min(pasted.length, 5);
    digitInputRefs.current[nextIdx]?.focus();
  };

  const autofillDevOtp = () => {
    if (!devOtpPreview) return;
    const chars = devOtpPreview.split('').slice(0, 6);
    setOtpDigits(chars);
    setErrorMsg('');
  };


  return (
    <div className="login-screen-wrapper">
      <div className="login-card-container">
        {/* Back to product overview button */}
        {onBackToProduct && (
          <button type="button" onClick={onBackToProduct} className="login-back-btn">
            <ArrowLeft size={16} />
            <span>Back to Product Overview</span>
          </button>
        )}

        {/* Header */}
        <div className="login-brand-header">
          <div className="login-brand-icon">
            <Shield size={28} />
          </div>
          <div>
            <h1>NHIT Secure Enterprise Portal</h1>
            <p>National Highways Infra Trust • Ministry of Road Transport & Highways</p>
          </div>
        </div>

        {/* Portal Destination Selector */}
        <div className="portal-selector-tabs">
          <button
            type="button"
            className={`portal-tab ${portalMode === 'admin' ? 'active' : ''}`}
            onClick={() => setPortalMode('admin')}
          >
            <Shield size={16} />
            <div>
              <strong>Admin Portal</strong>
              <span>Full governance, approvals & financial ledger</span>
            </div>
          </button>
          <button
            type="button"
            className={`portal-tab ${portalMode === 'user' ? 'active' : ''}`}
            onClick={() => setPortalMode('user')}
          >
            <User size={16} />
            <div>
              <strong>Field User Portal</strong>
              <span>Project submissions, GreenNotes & expense claims</span>
            </div>
          </button>
        </div>

        {errorMsg && (
          <div className="login-error-banner">
            <span>{errorMsg}</span>
          </div>
        )}

        {infoMsg && (
          <div className="portal-badge-indicator" style={{ color: 'var(--cyan)' }}>
            <CheckCircle2 size={14} />
            <span>{infoMsg}</span>
          </div>
        )}

        {/* ================= STEP 1: CREDENTIALS ================= */}
        {loginStep === 'credentials' && (
          <>
            {/* Google SSO Button */}
            <button
              type="button"
              className="google-sso-btn"
              onClick={handleGoogleSSO}
              disabled={loading}
              title="Sign in with your verified Google enterprise workspace identity"
            >
              <svg width="18" height="18" viewBox="0 0 24 24">
                <path
                  fill="#4285F4"
                  d="M23.745 12.27c0-.7-.06-1.4-.19-2.07H12v4.51h6.6c-.29 1.52-1.14 2.82-2.4 3.68v3.05h3.88c2.27-2.09 3.66-5.17 3.66-9.17z"
                />
                <path
                  fill="#34A853"
                  d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.88-3.05c-1.08.72-2.45 1.16-4.05 1.16-3.12 0-5.77-2.1-6.72-4.93H1.25v3.15C3.26 21.36 7.34 24 12 24z"
                />
                <path
                  fill="#FBBC05"
                  d="M5.28 14.27c-.25-.72-.38-1.49-.38-2.27s.13-1.55.38-2.27V6.58H1.25C.45 8.18 0 9.99 0 12s.45 3.82 1.25 5.42l4.03-3.15z"
                />
                <path
                  fill="#EA4335"
                  d="M12 4.75c1.77 0 3.35.61 4.6 1.8l3.42-3.42C17.95 1.19 15.24 0 12 0 7.34 0 3.26 2.64 1.25 6.58l4.03 3.15c.95-2.83 3.6-4.98 6.72-4.98z"
                />
              </svg>
              <span>Continue with Google Enterprise SSO</span>
            </button>

            {/* Quick Demo Access Button */}
            <button
              type="button"
              className="quick-demo-login-btn"
              onClick={handleQuickDemoLogin}
              disabled={loading}
              style={{
                marginTop: '10px',
                width: '100%',
                padding: '11px 16px',
                background: 'linear-gradient(135deg, rgba(56, 189, 248, 0.12), rgba(99, 102, 241, 0.15))',
                border: '1px solid rgba(56, 189, 248, 0.5)',
                borderRadius: '8px',
                color: '#38bdf8',
                fontWeight: 600,
                fontSize: '13px',
                cursor: 'pointer',
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                gap: '8px',
                transition: 'all 0.2s ease',
              }}
            >
              <span>⚡ 1-Click Demo Sign-In (Super Admin)</span>
            </button>

            <div className="auth-divider">
              <span>OR AUTHENTICATE WITH NHIT CREDENTIALS</span>
            </div>

            {/* Auth Mode Tabs (Password vs Passwordless OTP) */}
            <div className="auth-mode-tabs">
              <button
                type="button"
                className={`auth-mode-tab ${authTab === 'password' ? 'active' : ''}`}
                onClick={() => setAuthTab('password')}
              >
                <Lock size={14} />
                <span>Password + 2FA</span>
              </button>
              <button
                type="button"
                className={`auth-mode-tab ${authTab === 'direct_otp' ? 'active' : ''}`}
                onClick={() => setAuthTab('direct_otp')}
              >
                <Smartphone size={14} />
                <span>Passwordless OTP</span>
              </button>
            </div>

            {/* Form */}
            <form onSubmit={handleLoginSubmit} className="login-form">
              <div className="form-group">
                <label>
                  <Mail size={14} /> Registered NHIT Email
                </label>
                <input
                  type="email"
                  required
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  placeholder="e.g. admin@nhit.com"
                  className="login-input"
                />
              </div>

              {authTab === 'password' && (
                <div className="form-group">
                  <label>
                    <Lock size={14} /> Account Password
                  </label>
                  <input
                    type="password"
                    required
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    placeholder="••••••••••••"
                    className="login-input"
                  />
                </div>
              )}

              {/* 2FA Toggle for Password Flow */}
              {authTab === 'password' && (
                <label style={{ display: 'flex', alignItems: 'center', gap: '8px', cursor: 'pointer', fontSize: '0.78rem', color: 'var(--text-muted)' }}>
                  <input
                    type="checkbox"
                    checked={enable2FA}
                    onChange={(e) => setEnable2FA(e.target.checked)}
                    style={{ accentColor: 'var(--cyan)' }}
                  />
                  <span>Enforce Two-Factor (2-Way) OTP Security Challenge</span>
                </label>
              )}

              <div className="portal-badge-indicator">
                <CheckCircle2 size={14} />
                <span>
                  Targeting: <strong>{portalMode === 'admin' ? 'Executive Admin Console' : 'Field Officer Hub'}</strong>
                </span>
              </div>

              <button type="submit" className="login-submit-btn" disabled={loading}>
                {loading ? (
                  <span className="spinner" style={{ width: 18, height: 18 }}></span>
                ) : (
                  <>
                    <span>
                      {authTab === 'password'
                        ? enable2FA
                          ? 'Verify & Request 2FA Code'
                          : `Enter ${portalMode === 'admin' ? 'Admin Portal' : 'User Portal'}`
                        : 'Send 6-Digit Login Code'}
                    </span>
                    <ArrowRight size={16} />
                  </>
                )}
              </button>
            </form>
          </>
        )}

        {/* ================= STEP 2: 2FA OTP VERIFICATION ================= */}
        {loginStep === 'otp_verify' && (
          <div className="otp-verify-card">
            <div className="otp-instruction-header">
              <h3>
                <Shield size={20} className="text-cyan" />
                Two-Factor Security Verification
              </h3>
              <p>
                A 6-digit cryptographic verification code has been issued for{' '}
                <span className="otp-email-target">{email}</span>.
              </p>
            </div>

            {/* Split 6-Digit Input Group */}
            <div className="otp-input-group">
              {otpDigits.map((digit, idx) => (
                <input
                  key={idx}
                  ref={(el) => {
                    digitInputRefs.current[idx] = el;
                  }}
                  type="text"
                  maxLength={1}
                  value={digit}
                  onChange={(e) => handleDigitChange(idx, e.target.value)}
                  onKeyDown={(e) => handleDigitKeyDown(idx, e)}
                  onPaste={handleDigitPaste}
                  className="otp-digit-box"
                  autoFocus={idx === 0}
                  inputMode="numeric"
                />
              ))}
            </div>

            {/* OTP Meta: Timer and Resend */}
            <div className="otp-meta-row">
              <div className="otp-timer">
                <Clock size={14} />
                <span>Code expires in: {formatTimer(otpCountdown)}</span>
              </div>
              <button
                type="button"
                className="otp-resend-link"
                disabled={otpCountdown > 240 || loading}
                onClick={handleLoginSubmit}
              >
                <RefreshCw size={12} style={{ display: 'inline', marginRight: 4 }} />
                Resend Code
              </button>
            </div>

            {/* Dev Sandbox Quick-Fill Helper */}
            {devOtpPreview && (
              <div className="dev-otp-sandbox-pill" onClick={autofillDevOtp} title="Click to autofill generated code for fast testing">
                <span>⚡ Sandbox Generated Code: <code>{devOtpPreview}</code></span>
                <span style={{ textDecoration: 'underline', fontWeight: 600 }}>Click to Fill</span>
              </div>
            )}

            <button
              type="button"
              className="login-submit-btn"
              onClick={handleVerifyOTP}
              disabled={loading || otpDigits.join('').length < 6}
            >
              {loading ? (
                <span className="spinner" style={{ width: 18, height: 18 }}></span>
              ) : (
                <>
                  <span>Verify Code & Complete Sign-In</span>
                  <ArrowRight size={16} />
                </>
              )}
            </button>

            <button
              type="button"
              onClick={() => {
                setLoginStep('credentials');
                setErrorMsg('');
              }}
              className="login-back-btn"
              style={{ justifyContent: 'center' }}
            >
              <ArrowLeft size={14} />
              <span>Use different account / Go back</span>
            </button>
          </div>
        )}

        {/* ================= STEP 3: MANDATORY PASSWORD CHANGE (OPEN-03) ================= */}
        {loginStep === 'force_change_password' && (
          <div className="otp-verification-card">
            <div className="otp-badge-row">
              <KeyRound size={20} color="var(--primary-color)" />
              <h3 style={{ margin: 0 }}>Mandatory Password Update</h3>
            </div>
            <p className="otp-explainer">
              Initial or temporary credentials are in use for <strong>{email}</strong>. In compliance with NHIT IT Security Policy, please configure a new permanent password.
            </p>

            <form onSubmit={handleChangePassword} className="login-form">
              <div className="form-group">
                <label>
                  <Lock size={14} /> New Permanent Password (min. 8 characters)
                </label>
                <input
                  type="password"
                  required
                  minLength={8}
                  value={newPassword}
                  onChange={(e) => setNewPassword(e.target.value)}
                  placeholder="••••••••••••"
                  className="login-input"
                  autoComplete="new-password"
                  autoFocus
                />
              </div>

              <div className="form-group">
                <label>
                  <Lock size={14} /> Confirm Permanent Password
                </label>
                <input
                  type="password"
                  required
                  minLength={8}
                  value={confirmPassword}
                  onChange={(e) => setConfirmPassword(e.target.value)}
                  placeholder="••••••••••••"
                  className="login-input"
                  autoComplete="new-password"
                />
              </div>

              <button
                type="submit"
                className="login-submit-btn"
                disabled={loading || !newPassword || !confirmPassword}
              >
                {loading ? (
                  <span className="spinner" style={{ width: 18, height: 18 }}></span>
                ) : (
                  <>
                    <span>Update Password & Enter Portal</span>
                    <ArrowRight size={16} />
                  </>
                )}
              </button>

              <button
                type="button"
                onClick={() => {
                  setLoginStep('credentials');
                  setErrorMsg('');
                  setNewPassword('');
                  setConfirmPassword('');
                }}
                className="login-back-btn"
                style={{ justifyContent: 'center' }}
              >
                <ArrowLeft size={14} />
                <span>Cancel & Return to Login</span>
              </button>
            </form>
          </div>
        )}


        {/* Footer */}
        <div className="login-footer">
          <span>Protected by HMAC-SHA256 PyJWT + PostgreSQL Bcrypt Security Layer</span>
          <div className="login-links">
            <a href="http://localhost:8080/swagger" target="_blank" rel="noreferrer">
              Gateway Swagger UI (Port 8080)
            </a>
          </div>
        </div>
      </div>
    </div>
  );
};
