import React, { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  CardHeader,
  Paper,
  List,
  ListItem,
  ListItemText,
  Chip,
  Alert,
  Skeleton,
  Tabs,
  Tab,
} from "@mui/material";
import {
  People as PeopleIcon,
  ViewModule as ViewModuleIcon,
  Security as RolesIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
} from "@mui/icons-material";
import dashboardService, {
  DashboardOverview,
} from "../services/dashboardService";
import { useAuth } from "../contexts/AuthContext";
import UserInsightsComponent from "../components/Dashboard/UserInsights";
import CollectionInsightsComponent from "../components/Dashboard/CollectionInsights";

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: React.ReactElement;
  color: "primary" | "secondary" | "success" | "warning" | "info";
}

const StatsCard: React.FC<StatsCardProps> = ({ title, value, icon, color }) => (
  <Card sx={{ height: "100%" }}>
    <CardContent>
      <Box
        sx={{
          display: "flex",
          alignItems: "center",
          justifyContent: "space-between",
        }}
      >
        <Box>
          <Typography color="textSecondary" gutterBottom variant="overline">
            {title}
          </Typography>
          <Typography variant="h4" component="div">
            {value}
          </Typography>
        </Box>
        <Box
          sx={{
            backgroundColor: `${color}.light`,
            borderRadius: "50%",
            display: "flex",
            height: 56,
            justifyContent: "center",
            width: 56,
            alignItems: "center",
          }}
        >
          {React.cloneElement(icon, { color })}
        </Box>
      </Box>
    </CardContent>
  </Card>
);

const Dashboard: React.FC = () => {
  const { isAdmin } = useAuth();
  const [dashboardData, setDashboardData] = useState<DashboardOverview | null>(
    null
  );
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [currentTab, setCurrentTab] = useState(0);

  const handleTabChange = (event: React.SyntheticEvent, newValue: number) => {
    setCurrentTab(newValue);
  };

  useEffect(() => {
    const fetchDashboardData = async () => {
      // Don't fetch data if user is not admin
      // if (!isAdmin) {
      //   setLoading(false);
      //   return;
      // }

      try {
        setLoading(true);
        setError(null);
        const data = await dashboardService.getDashboardOverview();
        setDashboardData(data);
      } catch (err: any) {
        console.error("Error fetching dashboard data:", err);
        setError(err.response?.data?.error || "Failed to load dashboard data");
      } finally {
        setLoading(false);
      }
    };

    fetchDashboardData();
  }, [isAdmin]);

  const formatActivityTime = (timestamp: string): string => {
    const now = new Date();
    const activityTime = new Date(timestamp);
    const diffMs = now.getTime() - activityTime.getTime();
    const diffMinutes = Math.floor(diffMs / 60000);
    const diffHours = Math.floor(diffMs / 3600000);
    const diffDays = Math.floor(diffMs / 86400000);

    if (diffMinutes < 1) return "Just now";
    if (diffMinutes < 60)
      return `${diffMinutes} minute${diffMinutes !== 1 ? "s" : ""} ago`;
    if (diffHours < 24)
      return `${diffHours} hour${diffHours !== 1 ? "s" : ""} ago`;
    return `${diffDays} day${diffDays !== 1 ? "s" : ""} ago`;
  };

  const getActivityColor = (
    action: string
  ):
    | "default"
    | "primary"
    | "secondary"
    | "error"
    | "info"
    | "success"
    | "warning" => {
    switch (action.toLowerCase()) {
      case "create":
        return "success";
      case "update":
        return "info";
      case "delete":
        return "error";
      case "login":
        return "primary";
      default:
        return "default";
    }
  };

  // Check admin access
  // if (!isAdmin) {
  //   return (
  //     <Box>
  //       <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
  //         Dashboard
  //       </Typography>
  //       <Alert severity="error">
  //         Access denied. Administrator privileges required to view this page.
  //       </Alert>
  //     </Box>
  //   );
  // }

  if (loading) {
    return (
      <Box>
        <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
          Dashboard
        </Typography>
        <Grid container spacing={3} sx={{ mb: 3 }}>
          {[1, 2, 3, 4].map((index) => (
            <Grid item xs={12} sm={6} md={3} key={index}>
              <Card sx={{ height: "100%" }}>
                <CardContent>
                  <Box
                    sx={{
                      display: "flex",
                      alignItems: "center",
                      justifyContent: "space-between",
                    }}
                  >
                    <Box sx={{ flex: 1 }}>
                      <Skeleton variant="text" width="60%" />
                      <Skeleton variant="text" width="40%" height={40} />
                    </Box>
                    <Skeleton variant="circular" width={56} height={56} />
                  </Box>
                </CardContent>
              </Card>
            </Grid>
          ))}
        </Grid>
        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Paper sx={{ p: 3 }}>
              <Skeleton variant="text" width="30%" height={32} />
              <Skeleton variant="rectangular" height={100} sx={{ mt: 2 }} />
            </Paper>
          </Grid>
          <Grid item xs={12} md={4}>
            <Card>
              <CardHeader title={<Skeleton variant="text" width="50%" />} />
              <CardContent>
                <Skeleton variant="rectangular" height={200} />
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      </Box>
    );
  }

  if (error) {
    return (
      <Box>
        <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
          Dashboard
        </Typography>
        <Alert severity="error" sx={{ mb: 3 }}>
          {error}
        </Alert>
      </Box>
    );
  }

  if (!dashboardData) {
    return (
      <Box>
        <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
          Dashboard
        </Typography>
        <Alert severity="info">No dashboard data available</Alert>
      </Box>
    );
  }

  const stats = [
    {
      title: "Total Users",
      value: dashboardData.system_stats.total_users,
      icon: <PeopleIcon />,
      color: "primary" as const,
    },
    {
      title: "Active Users",
      value: dashboardData.system_stats.active_users,
      icon: <PeopleIcon />,
      color: "success" as const,
    },
    {
      title: "Collections",
      value: dashboardData.system_stats.total_collections,
      icon: <ViewModuleIcon />,
      color: "secondary" as const,
    },
    {
      title: "Roles",
      value: dashboardData.system_stats.total_roles,
      icon: <RolesIcon />,
      color: "info" as const,
    },
  ];

  return (
    <Box>
      <Typography variant="h4" sx={{ mb: 3, fontWeight: "bold" }}>
        Dashboard
      </Typography>

      <Grid container spacing={3} sx={{ mb: 3 }}>
        {stats.map((stat, index) => (
          <Grid item xs={12} sm={6} md={3} key={index}>
            <StatsCard {...stat} />
          </Grid>
        ))}
      </Grid>

      <Paper sx={{ mb: 3 }}>
        <Tabs
          value={currentTab}
          onChange={handleTabChange}
          sx={{ borderBottom: 1, borderColor: "divider" }}
        >
          <Tab label="System Overview" />
          <Tab label="User Analytics" />
          <Tab label="Collection Insights" />
        </Tabs>
      </Paper>

      {currentTab === 0 && (
        <Grid container spacing={3}>
          <Grid item xs={12} md={8}>
            <Paper sx={{ p: 3 }}>
              <Typography variant="h6" sx={{ mb: 2 }}>
                System Status
              </Typography>
              <Box sx={{ mb: 2 }}>
                <Typography
                  variant="body2"
                  color="textSecondary"
                  sx={{ mb: 1 }}
                >
                  Database Connection
                </Typography>
                <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                  {dashboardData.system_health.database_connected ? (
                    <CheckCircleIcon color="success" />
                  ) : (
                    <ErrorIcon color="error" />
                  )}
                  <Typography
                    variant="body2"
                    color={
                      dashboardData.system_health.database_connected
                        ? "success.main"
                        : "error.main"
                    }
                  >
                    {dashboardData.system_health.database_connected
                      ? "Connected"
                      : "Disconnected"}
                  </Typography>
                </Box>
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography
                  variant="body2"
                  color="textSecondary"
                  sx={{ mb: 1 }}
                >
                  Server Version
                </Typography>
                <Typography variant="body2">
                  {dashboardData.system_health.version}
                </Typography>
              </Box>
              <Box sx={{ mb: 2 }}>
                <Typography
                  variant="body2"
                  color="textSecondary"
                  sx={{ mb: 1 }}
                >
                  Server Uptime
                </Typography>
                <Typography variant="body2">
                  {dashboardData.system_health.server_uptime}
                </Typography>
              </Box>
              {dashboardData.system_health.last_backup && (
                <Box>
                  <Typography
                    variant="body2"
                    color="textSecondary"
                    sx={{ mb: 1 }}
                  >
                    Last Backup
                  </Typography>
                  <Typography variant="body2">
                    {new Date(
                      dashboardData.system_health.last_backup
                    ).toLocaleString()}
                  </Typography>
                </Box>
              )}
            </Paper>
          </Grid>

          <Grid item xs={12} md={4}>
            <Card>
              <CardHeader title="Recent Activity" />
              <CardContent sx={{ pt: 0 }}>
                <List dense>
                  {dashboardData.recent_activity &&
                  dashboardData.recent_activity.length > 0 ? (
                    dashboardData.recent_activity.map((activity) => (
                      <ListItem key={activity.id} sx={{ px: 0 }}>
                        <ListItemText
                          primary={
                            <Box
                              sx={{
                                display: "flex",
                                alignItems: "center",
                                gap: 1,
                              }}
                            >
                              <Typography variant="body2">
                                {activity.action}
                                {activity.collection &&
                                  ` in ${activity.collection}`}
                              </Typography>
                              <Chip
                                label={activity.action}
                                size="small"
                                color={getActivityColor(activity.action)}
                                sx={{ fontSize: "0.7rem", height: "20px" }}
                              />
                            </Box>
                          }
                          secondary={
                            <Box>
                              <Typography
                                variant="caption"
                                color="textSecondary"
                              >
                                by {activity.user_name || "System"}
                              </Typography>
                              <br />
                              <Typography
                                variant="caption"
                                color="textSecondary"
                              >
                                {formatActivityTime(activity.timestamp)}
                              </Typography>
                              {activity.comment && (
                                <>
                                  <br />
                                  <Typography
                                    variant="caption"
                                    color="textSecondary"
                                  >
                                    {activity.comment}
                                  </Typography>
                                </>
                              )}
                            </Box>
                          }
                        />
                      </ListItem>
                    ))
                  ) : (
                    <ListItem sx={{ px: 0 }}>
                      <ListItemText
                        primary={
                          <Typography variant="body2" color="textSecondary">
                            No recent activity
                          </Typography>
                        }
                      />
                    </ListItem>
                  )}
                </List>
              </CardContent>
            </Card>
          </Grid>
        </Grid>
      )}

      {currentTab === 1 && (
        <UserInsightsComponent data={dashboardData.user_insights} />
      )}

      {currentTab === 2 && (
        <CollectionInsightsComponent data={dashboardData.collection_metrics} />
      )}
    </Box>
  );
};

export default Dashboard;
