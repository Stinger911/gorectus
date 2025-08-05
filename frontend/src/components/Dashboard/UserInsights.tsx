import React from "react";
import {
  Box,
  Typography,
  Grid,
  Card,
  CardContent,
  CardHeader,
  List,
  ListItem,
  ListItemText,
  ListItemAvatar,
  Avatar,
  Chip,
  Divider,
} from "@mui/material";
import {
  Person as PersonIcon,
  Group as GroupIcon,
  TrendingUp as TrendingUpIcon,
} from "@mui/icons-material";
import { UserInsights, UserSummary } from "../../services/dashboardService";

interface UserInsightsComponentProps {
  data: UserInsights;
}

const UserInsightsComponent: React.FC<UserInsightsComponentProps> = ({
  data,
}) => {
  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString();
  };

  const getStatusColor = (
    status: string
  ):
    | "default"
    | "primary"
    | "secondary"
    | "error"
    | "info"
    | "success"
    | "warning" => {
    switch (status.toLowerCase()) {
      case "active":
        return "success";
      case "inactive":
        return "warning";
      case "invited":
        return "info";
      default:
        return "default";
    }
  };

  const getRoleColor = (
    role: string
  ):
    | "default"
    | "primary"
    | "secondary"
    | "error"
    | "info"
    | "success"
    | "warning" => {
    if (!role) return "default";
    switch (role.toLowerCase()) {
      case "administrator":
        return "error";
      case "editor":
        return "primary";
      case "user":
        return "secondary";
      default:
        return "default";
    }
  };

  return (
    <Grid container spacing={3}>
      {/* User Status Distribution */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader
            title="Users by Status"
            avatar={<GroupIcon color="primary" />}
          />
          <CardContent>
            <List dense>
              {Object.entries(data.users_by_status).map(([status, count]) => (
                <ListItem key={status} sx={{ px: 0 }}>
                  <ListItemText
                    primary={
                      <Box
                        sx={{
                          display: "flex",
                          justifyContent: "space-between",
                          alignItems: "center",
                        }}
                      >
                        <Typography
                          variant="body2"
                          sx={{ textTransform: "capitalize" }}
                        >
                          {status}
                        </Typography>
                        <Chip
                          label={count.toString()}
                          size="small"
                          color={getStatusColor(status)}
                        />
                      </Box>
                    }
                  />
                </ListItem>
              ))}
            </List>
          </CardContent>
        </Card>
      </Grid>

      {/* User Role Distribution */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader
            title="Users by Role"
            avatar={<PersonIcon color="primary" />}
          />
          <CardContent>
            <List dense>
              {Object.entries(data.users_by_role).map(([role, count]) => (
                <ListItem key={role} sx={{ px: 0 }}>
                  <ListItemText
                    primary={
                      <Box
                        sx={{
                          display: "flex",
                          justifyContent: "space-between",
                          alignItems: "center",
                        }}
                      >
                        <Typography variant="body2">{role}</Typography>
                        <Chip
                          label={count.toString()}
                          size="small"
                          color={getRoleColor(role)}
                        />
                      </Box>
                    }
                  />
                </ListItem>
              ))}
            </List>
          </CardContent>
        </Card>
      </Grid>

      {/* Growth Metrics */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader
            title="User Growth"
            avatar={<TrendingUpIcon color="primary" />}
          />
          <CardContent>
            <Box sx={{ mb: 2 }}>
              <Typography variant="body2" color="textSecondary">
                New Users This Week
              </Typography>
              <Typography variant="h4" color="primary">
                {data.new_users_this_week}
              </Typography>
            </Box>
            <Divider sx={{ my: 2 }} />
            <Box>
              <Typography variant="body2" color="textSecondary">
                New Users This Month
              </Typography>
              <Typography variant="h4" color="secondary">
                {data.new_users_this_month}
              </Typography>
            </Box>
          </CardContent>
        </Card>
      </Grid>

      {/* Recent Registrations */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader title="Recent Registrations" />
          <CardContent sx={{ pt: 0 }}>
            <List dense>
              {data.recent_registrations.length > 0 ? (
                data.recent_registrations.map((user: UserSummary) => (
                  <ListItem key={user.id} sx={{ px: 0 }}>
                    <ListItemAvatar>
                      <Avatar sx={{ bgcolor: "primary.main" }}>
                        {user.first_name.charAt(0)}
                        {user.last_name.charAt(0)}
                      </Avatar>
                    </ListItemAvatar>
                    <ListItemText
                      primary={`${user.first_name} ${user.last_name}`}
                      secondary={
                        <Box>
                          <Typography variant="caption" color="textSecondary">
                            {user.email}
                          </Typography>
                          <br />
                          <Typography variant="caption" color="textSecondary">
                            Registered: {formatDate(user.created_at)}
                          </Typography>
                          <br />
                          <Chip
                            label={user.role_name}
                            size="small"
                            color={getRoleColor(user.role_name)}
                            sx={{ mr: 1, mt: 0.5 }}
                          />
                          <Chip
                            label={user.status}
                            size="small"
                            color={getStatusColor(user.status)}
                            sx={{ mt: 0.5 }}
                          />
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
                        No recent registrations
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
  );
};

export default UserInsightsComponent;
