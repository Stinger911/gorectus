import React from "react";
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
  LinearProgress,
} from "@mui/material";
import {
  People as PeopleIcon,
  ViewModule as CollectionsIcon,
  Storage as StorageIcon,
  TrendingUp as TrendingUpIcon,
} from "@mui/icons-material";

interface StatsCardProps {
  title: string;
  value: string | number;
  icon: React.ReactElement;
  color: "primary" | "secondary" | "success" | "warning";
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
  const stats = [
    {
      title: "Total Users",
      value: 2,
      icon: <PeopleIcon />,
      color: "primary" as const,
    },
    {
      title: "Collections",
      value: 5,
      icon: <CollectionsIcon />,
      color: "secondary" as const,
    },
    {
      title: "Storage Used",
      value: "1.2 GB",
      icon: <StorageIcon />,
      color: "success" as const,
    },
    {
      title: "API Calls",
      value: "12.5K",
      icon: <TrendingUpIcon />,
      color: "warning" as const,
    },
  ];

  const recentActivity = [
    {
      action: "User login",
      user: "Admin User",
      time: "2 minutes ago",
      status: "success",
    },
    {
      action: "Collection created",
      user: "Admin User",
      time: "1 hour ago",
      status: "info",
    },
    {
      action: "Data updated",
      user: "Admin User",
      time: "3 hours ago",
      status: "warning",
    },
    {
      action: "User registered",
      user: "John Doe",
      time: "1 day ago",
      status: "success",
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

      <Grid container spacing={3}>
        <Grid item xs={12} md={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" sx={{ mb: 2 }}>
              System Status
            </Typography>
            <Box sx={{ mb: 2 }}>
              <Typography variant="body2" color="textSecondary" sx={{ mb: 1 }}>
                Database Connection
              </Typography>
              <LinearProgress
                variant="determinate"
                value={100}
                color="success"
                sx={{ mb: 1 }}
              />
              <Typography variant="caption" color="success.main">
                Connected
              </Typography>
            </Box>
            <Box sx={{ mb: 2 }}>
              <Typography variant="body2" color="textSecondary" sx={{ mb: 1 }}>
                Server Load
              </Typography>
              <LinearProgress
                variant="determinate"
                value={35}
                color="primary"
                sx={{ mb: 1 }}
              />
              <Typography variant="caption">35% CPU Usage</Typography>
            </Box>
            <Box>
              <Typography variant="body2" color="textSecondary" sx={{ mb: 1 }}>
                Memory Usage
              </Typography>
              <LinearProgress
                variant="determinate"
                value={68}
                color="warning"
                sx={{ mb: 1 }}
              />
              <Typography variant="caption" color="warning.main">
                68% Memory Used
              </Typography>
            </Box>
          </Paper>
        </Grid>

        <Grid item xs={12} md={4}>
          <Card>
            <CardHeader title="Recent Activity" />
            <CardContent sx={{ pt: 0 }}>
              <List dense>
                {recentActivity.map((activity, index) => (
                  <ListItem key={index} sx={{ px: 0 }}>
                    <ListItemText
                      primary={
                        <Box
                          sx={{ display: "flex", alignItems: "center", gap: 1 }}
                        >
                          <Typography variant="body2">
                            {activity.action}
                          </Typography>
                          <Chip
                            label={activity.status}
                            size="small"
                            color={activity.status as any}
                            sx={{ fontSize: "0.7rem", height: "20px" }}
                          />
                        </Box>
                      }
                      secondary={
                        <Box>
                          <Typography variant="caption" color="textSecondary">
                            by {activity.user}
                          </Typography>
                          <br />
                          <Typography variant="caption" color="textSecondary">
                            {activity.time}
                          </Typography>
                        </Box>
                      }
                    />
                  </ListItem>
                ))}
              </List>
            </CardContent>
          </Card>
        </Grid>
      </Grid>
    </Box>
  );
};

export default Dashboard;
