import React, { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Paper,
  Tabs,
  Tab,
  TextField,
  Button,
  Switch,
  FormControlLabel,
  Alert,
  Divider,
  CircularProgress,
} from "@mui/material";
import {
  Save as SaveIcon,
  Refresh as RefreshIcon,
  Security as SecurityIcon,
  Storage as StorageIcon,
  Mail as MailIcon,
} from "@mui/icons-material";
import { useAuth } from "../contexts/AuthContext";
import settingsService, {
  SystemPreferences,
  UpdateSettingsRequest,
} from "../services/settingsService";

interface TabPanelProps {
  children?: React.ReactNode;
  index: number;
  value: number;
}

function TabPanel(props: TabPanelProps) {
  const { children, value, index, ...other } = props;

  return (
    <div
      role="tabpanel"
      hidden={value !== index}
      id={`settings-tabpanel-${index}`}
      aria-labelledby={`settings-tab-${index}`}
      {...other}
    >
      {value === index && <Box sx={{ p: 3 }}>{children}</Box>}
    </div>
  );
}

const Settings: React.FC = () => {
  const { isAdmin } = useAuth();
  const [tabValue, setTabValue] = useState(0);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [settings, setSettings] = useState<SystemPreferences | null>(null);

  // Form state
  const [siteName, setSiteName] = useState("");
  const [siteDescription, setSiteDescription] = useState("");
  const [allowRegistration, setAllowRegistration] = useState(false);
  const [maintenanceMode, setMaintenanceMode] = useState(false);

  // Database Settings (read-only)
  const [dbHost, setDbHost] = useState("");
  const [dbPort, setDbPort] = useState("");
  const [dbName, setDbName] = useState("");
  const [dbUser, setDbUser] = useState("");

  // Email Settings
  const [smtpHost, setSmtpHost] = useState("");
  const [smtpPort, setSmtpPort] = useState("");
  const [smtpUser, setSmtpUser] = useState("");
  const [smtpFromEmail, setSmtpFromEmail] = useState("");
  const [emailEnabled, setEmailEnabled] = useState(false);

  // Security Settings
  const [jwtSecret, setJwtSecret] = useState("");
  const [sessionTimeout, setSessionTimeout] = useState("");
  const [passwordMinLength, setPasswordMinLength] = useState("");
  const [requireTwoFactor, setRequireTwoFactor] = useState(false);

  // Load settings on component mount
  useEffect(() => {
    if (!isAdmin) return;

    const loadSettings = async () => {
      try {
        setLoading(true);
        setError(null);
        const settingsData = await settingsService.getSettings();
        setSettings(settingsData);
        populateForm(settingsData);
      } catch (err: any) {
        console.error("Error loading settings:", err);
        setError(err.response?.data?.error || "Failed to load settings");
      } finally {
        setLoading(false);
      }
    };

    loadSettings();
  }, [isAdmin]);

  const populateForm = (settingsData: SystemPreferences) => {
    setSiteName(settingsData.site_name);
    setSiteDescription(settingsData.site_description);
    setAllowRegistration(settingsData.allow_registration);
    setMaintenanceMode(settingsData.maintenance_mode);

    setDbHost(settingsData.database_host);
    setDbPort(settingsData.database_port);
    setDbName(settingsData.database_name);
    setDbUser(settingsData.database_user);

    setSmtpHost(settingsData.smtp_host);
    setSmtpPort(settingsData.smtp_port);
    setSmtpUser(settingsData.smtp_user);
    setSmtpFromEmail(settingsData.smtp_from_email);
    setEmailEnabled(settingsData.email_enabled);

    setSessionTimeout(settingsData.session_timeout.toString());
    setPasswordMinLength(settingsData.password_min_length.toString());
    setRequireTwoFactor(settingsData.require_two_factor);

    // Don't populate JWT secret for security
    setJwtSecret("");
  };

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleSave = async () => {
    if (!settings) return;

    try {
      setSaving(true);
      setError(null);

      const updateData: UpdateSettingsRequest = {};

      // Only include changed values
      if (siteName !== settings.site_name) updateData.site_name = siteName;
      if (siteDescription !== settings.site_description)
        updateData.site_description = siteDescription;
      if (allowRegistration !== settings.allow_registration)
        updateData.allow_registration = allowRegistration;
      if (maintenanceMode !== settings.maintenance_mode)
        updateData.maintenance_mode = maintenanceMode;

      if (smtpHost !== settings.smtp_host) updateData.smtp_host = smtpHost;
      if (smtpPort !== settings.smtp_port) updateData.smtp_port = smtpPort;
      if (smtpUser !== settings.smtp_user) updateData.smtp_user = smtpUser;
      if (smtpFromEmail !== settings.smtp_from_email)
        updateData.smtp_from_email = smtpFromEmail;
      if (emailEnabled !== settings.email_enabled)
        updateData.email_enabled = emailEnabled;

      if (jwtSecret.trim() !== "") updateData.jwt_secret = jwtSecret;
      if (sessionTimeout !== settings.session_timeout.toString())
        updateData.session_timeout = parseInt(sessionTimeout);
      if (passwordMinLength !== settings.password_min_length.toString())
        updateData.password_min_length = parseInt(passwordMinLength);
      if (requireTwoFactor !== settings.require_two_factor)
        updateData.require_two_factor = requireTwoFactor;

      // Only make API call if there are changes
      if (Object.keys(updateData).length > 0) {
        const updatedSettings =
          await settingsService.updateSettings(updateData);
        setSettings(updatedSettings);
        populateForm(updatedSettings);
      }

      setSaveSuccess(true);
      setTimeout(() => setSaveSuccess(false), 3000);
    } catch (err: any) {
      console.error("Error saving settings:", err);
      setError(err.response?.data?.error || "Failed to save settings");
    } finally {
      setSaving(false);
    }
  };

  const handleTestConnection = async () => {
    try {
      setError(null);
      await settingsService.testDatabaseConnection();
      alert("Database connection successful!");
    } catch (err: any) {
      const message = err.response?.data?.error || "Database connection failed";
      setError(message);
      alert(message);
    }
  };

  const handleTestEmail = async () => {
    try {
      setError(null);
      await settingsService.testEmailConfiguration();
      alert("Test email sent successfully!");
    } catch (err: any) {
      const message = err.response?.data?.error || "Email test failed";
      setError(message);
      alert(message);
    }
  };

  const generateJwtSecret = () => {
    const chars =
      "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789";
    let result = "";
    for (let i = 0; i < 64; i++) {
      result += chars.charAt(Math.floor(Math.random() * chars.length));
    }
    setJwtSecret(result);
  };

  if (!isAdmin) {
    return (
      <Box>
        <Alert severity="error">
          Access denied. Administrator privileges required to view this page.
        </Alert>
      </Box>
    );
  }

  if (loading) {
    return (
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: "200px",
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
        Settings
      </Typography>

      {error && (
        <Alert severity="error" sx={{ mb: 2 }}>
          {error}
        </Alert>
      )}

      {saveSuccess && (
        <Alert severity="success" sx={{ mb: 2 }}>
          Settings saved successfully!
        </Alert>
      )}

      <Paper sx={{ width: "100%" }}>
        <Box sx={{ borderBottom: 1, borderColor: "divider" }}>
          <Tabs value={tabValue} onChange={handleTabChange}>
            <Tab label="General" />
            <Tab label="Database" />
            <Tab label="Email" />
            <Tab label="Security" />
          </Tabs>
        </Box>

        <TabPanel value={tabValue} index={0}>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
            <Typography
              variant="h6"
              sx={{ display: "flex", alignItems: "center", gap: 1 }}
            >
              <SecurityIcon />
              General Configuration
            </Typography>

            <TextField
              label="Site Name"
              value={siteName}
              onChange={(e) => setSiteName(e.target.value)}
              fullWidth
            />

            <TextField
              label="Site Description"
              value={siteDescription}
              onChange={(e) => setSiteDescription(e.target.value)}
              multiline
              rows={3}
              fullWidth
            />

            <Divider />

            <FormControlLabel
              control={
                <Switch
                  checked={allowRegistration}
                  onChange={(e) => setAllowRegistration(e.target.checked)}
                />
              }
              label="Allow User Registration"
            />

            <FormControlLabel
              control={
                <Switch
                  checked={maintenanceMode}
                  onChange={(e) => setMaintenanceMode(e.target.checked)}
                />
              }
              label="Maintenance Mode"
            />

            <Box sx={{ display: "flex", gap: 2, mt: 2 }}>
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handleSave}
                disabled={saving}
              >
                {saving ? "Saving..." : "Save Changes"}
              </Button>
            </Box>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={1}>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
            <Typography
              variant="h6"
              sx={{ display: "flex", alignItems: "center", gap: 1 }}
            >
              <StorageIcon />
              Database Configuration
            </Typography>

            <Box sx={{ display: "flex", gap: 2 }}>
              <TextField
                label="Host"
                value={dbHost}
                onChange={(e) => setDbHost(e.target.value)}
                fullWidth
              />
              <TextField
                label="Port"
                value={dbPort}
                onChange={(e) => setDbPort(e.target.value)}
                sx={{ width: "150px" }}
              />
            </Box>

            <TextField
              label="Database Name"
              value={dbName}
              onChange={(e) => setDbName(e.target.value)}
              fullWidth
            />

            <TextField
              label="Username"
              value={dbUser}
              onChange={(e) => setDbUser(e.target.value)}
              fullWidth
            />

            <Alert severity="info">
              Database password is managed through environment variables for
              security.
            </Alert>

            <Box sx={{ display: "flex", gap: 2, mt: 2 }}>
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handleSave}
                disabled={saving}
              >
                {saving ? "Saving..." : "Save Changes"}
              </Button>
              <Button
                variant="outlined"
                onClick={handleTestConnection}
                disabled={saving}
              >
                Test Connection
              </Button>
            </Box>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={2}>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
            <Typography
              variant="h6"
              sx={{ display: "flex", alignItems: "center", gap: 1 }}
            >
              <MailIcon />
              Email Configuration
            </Typography>

            <FormControlLabel
              control={
                <Switch
                  checked={emailEnabled}
                  onChange={(e) => setEmailEnabled(e.target.checked)}
                />
              }
              label="Enable Email Notifications"
            />

            <Box sx={{ display: "flex", gap: 2 }}>
              <TextField
                label="SMTP Host"
                value={smtpHost}
                onChange={(e) => setSmtpHost(e.target.value)}
                fullWidth
                disabled={!emailEnabled}
              />
              <TextField
                label="SMTP Port"
                value={smtpPort}
                onChange={(e) => setSmtpPort(e.target.value)}
                sx={{ width: "150px" }}
                disabled={!emailEnabled}
              />
            </Box>

            <TextField
              label="SMTP Username"
              value={smtpUser}
              onChange={(e) => setSmtpUser(e.target.value)}
              fullWidth
              disabled={!emailEnabled}
            />

            <TextField
              label="From Email Address"
              value={smtpFromEmail}
              onChange={(e) => setSmtpFromEmail(e.target.value)}
              fullWidth
              disabled={!emailEnabled}
            />

            <Alert severity="info">
              SMTP password is managed through environment variables for
              security.
            </Alert>

            <Box sx={{ display: "flex", gap: 2, mt: 2 }}>
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handleSave}
                disabled={!emailEnabled || saving}
              >
                {saving ? "Saving..." : "Save Changes"}
              </Button>
              <Button
                variant="outlined"
                onClick={handleTestEmail}
                disabled={!emailEnabled || saving}
              >
                Send Test Email
              </Button>
            </Box>
          </Box>
        </TabPanel>

        <TabPanel value={tabValue} index={3}>
          <Box sx={{ display: "flex", flexDirection: "column", gap: 3 }}>
            <Typography
              variant="h6"
              sx={{ display: "flex", alignItems: "center", gap: 1 }}
            >
              <SecurityIcon />
              Security Configuration
            </Typography>

            <Box>
              <TextField
                label="JWT Secret Key"
                value={jwtSecret}
                onChange={(e) => setJwtSecret(e.target.value)}
                fullWidth
                type="password"
                helperText="Used to sign JWT tokens. Keep this secret and secure."
              />
              <Button
                variant="outlined"
                size="small"
                onClick={generateJwtSecret}
                startIcon={<RefreshIcon />}
                sx={{ mt: 1 }}
              >
                Generate New Secret
              </Button>
            </Box>

            <TextField
              label="Session Timeout (hours)"
              value={sessionTimeout}
              onChange={(e) => setSessionTimeout(e.target.value)}
              type="number"
              sx={{ width: "200px" }}
            />

            <TextField
              label="Minimum Password Length"
              value={passwordMinLength}
              onChange={(e) => setPasswordMinLength(e.target.value)}
              type="number"
              sx={{ width: "200px" }}
            />

            <FormControlLabel
              control={
                <Switch
                  checked={requireTwoFactor}
                  onChange={(e) => setRequireTwoFactor(e.target.checked)}
                />
              }
              label="Require Two-Factor Authentication"
            />

            <Alert severity="warning">
              Changing security settings may affect all logged-in users. Use
              with caution.
            </Alert>

            <Box sx={{ display: "flex", gap: 2, mt: 2 }}>
              <Button
                variant="contained"
                startIcon={<SaveIcon />}
                onClick={handleSave}
                disabled={saving}
              >
                {saving ? "Saving..." : "Save Changes"}
              </Button>
            </Box>
          </Box>
        </TabPanel>
      </Paper>
    </Box>
  );
};

export default Settings;
