import React, { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  Grid,
  Card,
  CardContent,
  CardActions,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  FormControl,
  InputLabel,
  Select,
  MenuItem,
  Chip,
  Alert,
  CircularProgress,
  List,
  ListItem,
  ListItemText,
  ListItemSecondaryAction,
  IconButton,
  Accordion,
  AccordionSummary,
  AccordionDetails,
} from "@mui/material";
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  ViewList as ViewListIcon,
  ExpandMore as ExpandMoreIcon,
  Storage as StorageIcon,
} from "@mui/icons-material";

interface Collection {
  id: number;
  name: string;
  table_name: string;
  schema: CollectionField[];
  created_at: string;
  updated_at: string;
  item_count: number;
}

interface CollectionField {
  id: number;
  field_name: string;
  field_type: "string" | "number" | "boolean" | "date" | "text" | "json";
  required: boolean;
  unique: boolean;
  default_value: string | null;
}

interface CollectionFormData {
  name: string;
  table_name: string;
  schema: CollectionField[];
}

const Collections: React.FC = () => {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingCollection, setEditingCollection] = useState<Collection | null>(
    null
  );
  const [formData, setFormData] = useState<CollectionFormData>({
    name: "",
    table_name: "",
    schema: [],
  });

  // Mock data for now - will be replaced with API calls
  useEffect(() => {
    const mockCollections: Collection[] = [
      {
        id: 1,
        name: "Articles",
        table_name: "articles",
        schema: [
          {
            id: 1,
            field_name: "title",
            field_type: "string",
            required: true,
            unique: false,
            default_value: null,
          },
          {
            id: 2,
            field_name: "content",
            field_type: "text",
            required: true,
            unique: false,
            default_value: null,
          },
          {
            id: 3,
            field_name: "published",
            field_type: "boolean",
            required: false,
            unique: false,
            default_value: "false",
          },
        ],
        created_at: "2024-01-01T00:00:00Z",
        updated_at: "2024-01-15T10:30:00Z",
        item_count: 25,
      },
      {
        id: 2,
        name: "Products",
        table_name: "products",
        schema: [
          {
            id: 4,
            field_name: "name",
            field_type: "string",
            required: true,
            unique: false,
            default_value: null,
          },
          {
            id: 5,
            field_name: "price",
            field_type: "number",
            required: true,
            unique: false,
            default_value: null,
          },
          {
            id: 6,
            field_name: "description",
            field_type: "text",
            required: false,
            unique: false,
            default_value: null,
          },
        ],
        created_at: "2024-01-05T00:00:00Z",
        updated_at: "2024-01-10T15:20:00Z",
        item_count: 12,
      },
    ];

    setTimeout(() => {
      setCollections(mockCollections);
      setLoading(false);
    }, 1000);
  }, []);

  const handleOpenDialog = (collection?: Collection) => {
    if (collection) {
      setEditingCollection(collection);
      setFormData({
        name: collection.name,
        table_name: collection.table_name,
        schema: collection.schema,
      });
    } else {
      setEditingCollection(null);
      setFormData({
        name: "",
        table_name: "",
        schema: [],
      });
    }
    setDialogOpen(true);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditingCollection(null);
    setError(null);
  };

  const handleFormChange = (field: keyof CollectionFormData, value: any) => {
    setFormData((prev) => ({ ...prev, [field]: value }));

    // Auto-generate table name from collection name
    if (field === "name" && typeof value === "string") {
      const tableName = value
        .toLowerCase()
        .replace(/[^a-z0-9]/g, "_")
        .replace(/_+/g, "_");
      setFormData((prev) => ({ ...prev, table_name: tableName }));
    }
  };

  const addField = () => {
    const newField: CollectionField = {
      id: Date.now(),
      field_name: "",
      field_type: "string",
      required: false,
      unique: false,
      default_value: null,
    };
    setFormData((prev) => ({
      ...prev,
      schema: [...prev.schema, newField],
    }));
  };

  const updateField = (index: number, field: Partial<CollectionField>) => {
    setFormData((prev) => ({
      ...prev,
      schema: prev.schema.map((f, i) => (i === index ? { ...f, ...field } : f)),
    }));
  };

  const removeField = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      schema: prev.schema.filter((_, i) => i !== index),
    }));
  };

  const handleSubmit = async () => {
    setError(null);

    // Basic validation
    if (!formData.name || !formData.table_name) {
      setError("Please fill in collection name and table name");
      return;
    }

    if (formData.schema.length === 0) {
      setError("At least one field is required");
      return;
    }

    // Validate fields
    for (const field of formData.schema) {
      if (!field.field_name) {
        setError("All fields must have a name");
        return;
      }
    }

    try {
      // Mock API call - will be replaced with actual API
      if (editingCollection) {
        // Update existing collection
        const updatedCollection: Collection = {
          ...editingCollection,
          ...formData,
          updated_at: new Date().toISOString(),
        };
        setCollections((prev) =>
          prev.map((c) =>
            c.id === editingCollection.id ? updatedCollection : c
          )
        );
      } else {
        // Create new collection
        const newCollection: Collection = {
          id: Math.max(...collections.map((c) => c.id)) + 1,
          ...formData,
          created_at: new Date().toISOString(),
          updated_at: new Date().toISOString(),
          item_count: 0,
        };
        setCollections((prev) => [...prev, newCollection]);
      }

      handleCloseDialog();
    } catch (err) {
      setError("Failed to save collection");
    }
  };

  const handleDelete = async (collectionId: number) => {
    if (
      window.confirm(
        "Are you sure you want to delete this collection? This will also delete all data in the collection."
      )
    ) {
      // Mock API call - will be replaced with actual API
      setCollections((prev) => prev.filter((c) => c.id !== collectionId));
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const getFieldTypeColor = (type: string) => {
    switch (type) {
      case "string":
        return "primary";
      case "number":
        return "secondary";
      case "boolean":
        return "success";
      case "date":
        return "warning";
      case "text":
        return "info";
      case "json":
        return "error";
      default:
        return "default";
    }
  };

  if (loading) {
    return (
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          height: 400,
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  return (
    <Box>
      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          mb: 3,
        }}
      >
        <Typography variant="h4" sx={{ fontWeight: "bold" }}>
          Collections
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
        >
          Create Collection
        </Button>
      </Box>

      <Grid container spacing={3}>
        {collections.map((collection) => (
          <Grid item xs={12} md={6} lg={4} key={collection.id}>
            <Card
              sx={{ height: "100%", display: "flex", flexDirection: "column" }}
            >
              <CardContent sx={{ flexGrow: 1 }}>
                <Box
                  sx={{ display: "flex", alignItems: "center", gap: 1, mb: 2 }}
                >
                  <StorageIcon color="primary" />
                  <Typography variant="h6" component="div">
                    {collection.name}
                  </Typography>
                </Box>

                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{ mb: 1 }}
                >
                  Table: {collection.table_name}
                </Typography>

                <Typography
                  variant="body2"
                  color="text.secondary"
                  sx={{ mb: 2 }}
                >
                  {collection.item_count} items
                </Typography>

                <Accordion>
                  <AccordionSummary expandIcon={<ExpandMoreIcon />}>
                    <Typography variant="body2">
                      Schema ({collection.schema.length} fields)
                    </Typography>
                  </AccordionSummary>
                  <AccordionDetails>
                    <List dense>
                      {collection.schema.map((field, index) => (
                        <ListItem key={index} sx={{ px: 0 }}>
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
                                  {field.field_name}
                                </Typography>
                                {field.required && (
                                  <Chip
                                    label="Required"
                                    size="small"
                                    color="error"
                                  />
                                )}
                                {field.unique && (
                                  <Chip
                                    label="Unique"
                                    size="small"
                                    color="warning"
                                  />
                                )}
                              </Box>
                            }
                            secondary={
                              <Chip
                                label={field.field_type}
                                size="small"
                                color={
                                  getFieldTypeColor(field.field_type) as any
                                }
                              />
                            }
                          />
                        </ListItem>
                      ))}
                    </List>
                  </AccordionDetails>
                </Accordion>

                <Typography
                  variant="caption"
                  color="text.secondary"
                  sx={{ mt: 2, display: "block" }}
                >
                  Created: {formatDate(collection.created_at)}
                </Typography>
              </CardContent>

              <CardActions>
                <IconButton size="small" title="View Data">
                  <ViewListIcon />
                </IconButton>
                <IconButton
                  size="small"
                  onClick={() => handleOpenDialog(collection)}
                  title="Edit"
                >
                  <EditIcon />
                </IconButton>
                <IconButton
                  size="small"
                  onClick={() => handleDelete(collection.id)}
                  title="Delete"
                  color="error"
                >
                  <DeleteIcon />
                </IconButton>
              </CardActions>
            </Card>
          </Grid>
        ))}
      </Grid>

      <Dialog
        open={dialogOpen}
        onClose={handleCloseDialog}
        maxWidth="md"
        fullWidth
      >
        <DialogTitle>
          {editingCollection ? "Edit Collection" : "Create New Collection"}
        </DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
            <TextField
              label="Collection Name"
              value={formData.name}
              onChange={(e) => handleFormChange("name", e.target.value)}
              required
              fullWidth
            />
            <TextField
              label="Table Name"
              value={formData.table_name}
              onChange={(e) => handleFormChange("table_name", e.target.value)}
              required
              fullWidth
              helperText="Database table name (auto-generated from collection name)"
            />

            <Box sx={{ mt: 2 }}>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  mb: 2,
                }}
              >
                <Typography variant="h6">Schema Fields</Typography>
                <Button startIcon={<AddIcon />} onClick={addField}>
                  Add Field
                </Button>
              </Box>

              {formData.schema.map((field, index) => (
                <Paper key={field.id} sx={{ p: 2, mb: 2 }}>
                  <Box
                    sx={{ display: "flex", flexDirection: "column", gap: 2 }}
                  >
                    <Box sx={{ display: "flex", gap: 2, alignItems: "center" }}>
                      <TextField
                        label="Field Name"
                        value={field.field_name}
                        onChange={(e) =>
                          updateField(index, { field_name: e.target.value })
                        }
                        required
                        size="small"
                        sx={{ flex: 1 }}
                      />
                      <FormControl size="small" sx={{ minWidth: 120 }}>
                        <InputLabel>Type</InputLabel>
                        <Select
                          value={field.field_type}
                          label="Type"
                          onChange={(e) =>
                            updateField(index, {
                              field_type: e.target.value as any,
                            })
                          }
                        >
                          <MenuItem value="string">String</MenuItem>
                          <MenuItem value="number">Number</MenuItem>
                          <MenuItem value="boolean">Boolean</MenuItem>
                          <MenuItem value="date">Date</MenuItem>
                          <MenuItem value="text">Text</MenuItem>
                          <MenuItem value="json">JSON</MenuItem>
                        </Select>
                      </FormControl>
                      <IconButton
                        onClick={() => removeField(index)}
                        color="error"
                        size="small"
                      >
                        <DeleteIcon />
                      </IconButton>
                    </Box>

                    <Box sx={{ display: "flex", gap: 2 }}>
                      <FormControl size="small">
                        <InputLabel>Required</InputLabel>
                        <Select
                          value={field.required ? "yes" : "no"}
                          label="Required"
                          onChange={(e) =>
                            updateField(index, {
                              required: e.target.value === "yes",
                            })
                          }
                        >
                          <MenuItem value="no">No</MenuItem>
                          <MenuItem value="yes">Yes</MenuItem>
                        </Select>
                      </FormControl>

                      <FormControl size="small">
                        <InputLabel>Unique</InputLabel>
                        <Select
                          value={field.unique ? "yes" : "no"}
                          label="Unique"
                          onChange={(e) =>
                            updateField(index, {
                              unique: e.target.value === "yes",
                            })
                          }
                        >
                          <MenuItem value="no">No</MenuItem>
                          <MenuItem value="yes">Yes</MenuItem>
                        </Select>
                      </FormControl>

                      <TextField
                        label="Default Value"
                        value={field.default_value || ""}
                        onChange={(e) =>
                          updateField(index, {
                            default_value: e.target.value || null,
                          })
                        }
                        size="small"
                        sx={{ flex: 1 }}
                      />
                    </Box>
                  </Box>
                </Paper>
              ))}
            </Box>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button onClick={handleSubmit} variant="contained">
            {editingCollection ? "Update" : "Create"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Collections;
