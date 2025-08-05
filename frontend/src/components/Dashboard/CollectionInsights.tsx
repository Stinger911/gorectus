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
  ListItemIcon,
  Chip,
  Avatar,
} from "@mui/material";
import {
  ViewModule as ViewModuleIcon,
  Article as ArticleIcon,
  Settings as SettingsIcon,
  FolderOpen as FolderOpenIcon,
} from "@mui/icons-material";
import {
  CollectionMetrics,
  CollectionSummary,
} from "../../services/dashboardService";

interface CollectionInsightsComponentProps {
  data: CollectionMetrics;
}

const CollectionInsightsComponent: React.FC<
  CollectionInsightsComponentProps
> = ({ data }) => {
  const formatDate = (dateString: string): string => {
    return new Date(dateString).toLocaleDateString();
  };

  const getCollectionIcon = (collection: string) => {
    // Simple icon mapping based on collection name
    if (collection.includes("article") || collection.includes("post")) {
      return <ArticleIcon />;
    }
    if (collection.includes("setting") || collection.includes("config")) {
      return <SettingsIcon />;
    }
    return <ViewModuleIcon />;
  };

  const getTypeColor = (
    type: string
  ):
    | "default"
    | "primary"
    | "secondary"
    | "error"
    | "info"
    | "success"
    | "warning" => {
    switch (type.toLowerCase()) {
      case "content":
        return "primary";
      case "system":
        return "warning";
      case "ungrouped":
        return "secondary";
      default:
        return "default";
    }
  };

  return (
    <Grid container spacing={3}>
      {/* Collection Types Distribution */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader
            title="Collections by Type"
            avatar={<FolderOpenIcon color="primary" />}
          />
          <CardContent>
            <List dense>
              {Object.entries(data.collections_by_type).map(([type, count]) => (
                <ListItem key={type} sx={{ px: 0 }}>
                  <ListItemText
                    primary={
                      <Box
                        sx={{
                          display: "flex",
                          justifyContent: "space-between",
                          alignItems: "center",
                        }}
                      >
                        <Typography variant="body2">{type}</Typography>
                        <Chip
                          label={count.toString()}
                          size="small"
                          color={getTypeColor(type)}
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

      {/* Collection Summary */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader title="Collection Overview" />
          <CardContent>
            <Box sx={{ textAlign: "center" }}>
              <Typography variant="h3" color="primary" sx={{ mb: 1 }}>
                {data.total_collections}
              </Typography>
              <Typography variant="body2" color="textSecondary">
                Total Collections
              </Typography>
            </Box>
          </CardContent>
        </Card>
      </Grid>

      {/* Recent Collections */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader title="Recent Collections" />
          <CardContent sx={{ pt: 0 }}>
            <List dense>
              {data.recent_collections && data.recent_collections.length > 0 ? (
                data.recent_collections.map((collection: CollectionSummary) => (
                  <ListItem key={collection.collection} sx={{ px: 0 }}>
                    <ListItemIcon>
                      <Avatar
                        sx={{ bgcolor: "primary.light", width: 32, height: 32 }}
                      >
                        {getCollectionIcon(collection.collection)}
                      </Avatar>
                    </ListItemIcon>
                    <ListItemText
                      primary={collection.collection}
                      secondary={
                        <Box>
                          {collection.note && (
                            <>
                              <Typography
                                variant="caption"
                                color="textSecondary"
                              >
                                {collection.note}
                              </Typography>
                              <br />
                            </>
                          )}
                          <Typography variant="caption" color="textSecondary">
                            Created: {formatDate(collection.created_at)}
                          </Typography>
                          <br />
                          <Chip
                            label={`${collection.item_count} items`}
                            size="small"
                            color="info"
                            sx={{ mt: 0.5 }}
                          />
                          {collection.singleton && (
                            <Chip
                              label="Singleton"
                              size="small"
                              color="warning"
                              sx={{ ml: 1, mt: 0.5 }}
                            />
                          )}
                          {collection.hidden && (
                            <Chip
                              label="Hidden"
                              size="small"
                              color="secondary"
                              sx={{ ml: 1, mt: 0.5 }}
                            />
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
                        No recent collections
                      </Typography>
                    }
                  />
                </ListItem>
              )}
            </List>
          </CardContent>
        </Card>
      </Grid>

      {/* Most Active Collections */}
      <Grid item xs={12} md={6}>
        <Card>
          <CardHeader title="Most Active Collections" />
          <CardContent sx={{ pt: 0 }}>
            <List dense>
              {data.most_active_collections &&
              data.most_active_collections.length > 0 ? (
                data.most_active_collections.map(
                  (collection: CollectionSummary) => (
                    <ListItem key={collection.collection} sx={{ px: 0 }}>
                      <ListItemIcon>
                        <Avatar
                          sx={{
                            bgcolor: "secondary.light",
                            width: 32,
                            height: 32,
                          }}
                        >
                          {getCollectionIcon(collection.collection)}
                        </Avatar>
                      </ListItemIcon>
                      <ListItemText
                        primary={collection.collection}
                        secondary={
                          <Box>
                            {collection.note && (
                              <>
                                <Typography
                                  variant="caption"
                                  color="textSecondary"
                                >
                                  {collection.note}
                                </Typography>
                                <br />
                              </>
                            )}
                            <Chip
                              label={`${collection.item_count} items`}
                              size="small"
                              color="success"
                              sx={{ mt: 0.5 }}
                            />
                          </Box>
                        }
                      />
                    </ListItem>
                  )
                )
              ) : (
                <ListItem sx={{ px: 0 }}>
                  <ListItemText
                    primary={
                      <Typography variant="body2" color="textSecondary">
                        No active collections
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

export default CollectionInsightsComponent;
