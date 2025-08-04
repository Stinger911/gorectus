import React, { useState } from "react";
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
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
} from "@mui/material";
import {
  Save as SaveIcon,
  Refresh as RefreshIcon,
  Security as SecurityIcon,
  Storage as StorageIcon,
  Mail as MailIcon,
} from "@mui/icons-material";

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
  const [tabValue, setTabValue] = useState(0);
  const [saveSuccess, setSaveSuccess] = useState(false);

  // General Settings
  const [siteName, setSiteName] = useState("GoRectus");
  const [siteDescription, setSiteDescription] = useState(
    "A modern headless CMS built with Go"
  );
  const [allowRegistration, setAllowRegistration] = useState(false);
  const [maintenanceMode, setMaintenanceMode] = useState(false);

  // Database Settings
  const [dbHost, setDbHost] = useState("localhost");
  const [dbPort, setDbPort] = useState("5432");
  const [dbName, setDbName] = useState("gorectus");
  const [dbUser, setDbUser] = useState("gorectus");

  // Email Settings
  const [smtpHost, setSmtpHost] = useState("");
  const [smtpPort, setSmtpPort] = useState("587");
  const [smtpUser, setSmtpUser] = useState("");
  const [smtpFromEmail, setSmtpFromEmail] = useState("");
  const [emailEnabled, setEmailEnabled] = useState(false);

  // Security Settings
  const [jwtSecret, setJwtSecret] = useState("your-secret-key");
  const [sessionTimeout, setSessionTimeout] = useState("24");
  const [passwordMinLength, setPasswordMinLength] = useState("8");
  const [requireTwoFactor, setRequireTwoFactor] = useState(false);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setTabValue(newValue);
  };

  const handleSave = () => {
    // Mock save - in real app, this would make API calls
    setSaveSuccess(true);
    setTimeout(() => setSaveSuccess(false), 3000);
  };

  const handleTestConnection = () => {
    // Mock database connection test
    alert("Database connection successful!");
  };

  const handleTestEmail = () => {
    // Mock email test
    alert("Test email sent successfully!");
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

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
        Settings
      </Typography>

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
              >
                Save Changes
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
              >
                Save Changes
              </Button>
              <Button variant="outlined" onClick={handleTestConnection}>
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
                disabled={!emailEnabled}
              >
                Save Changes
              </Button>
              <Button
                variant="outlined"
                onClick={handleTestEmail}
                disabled={!emailEnabled}
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
              >
                Save Changes
              </Button>
            </Box>
          </Box>
        </TabPanel>
      </Paper>
    </Box>
  );
};

export default Settings;
