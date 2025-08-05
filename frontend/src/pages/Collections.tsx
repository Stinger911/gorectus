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
  IconButton,
} from "@mui/material";
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  ViewList as ViewListIcon,
  Storage as StorageIcon,
} from "@mui/icons-material";
import collectionsService, {
  Collection,
  CreateCollectionRequest,
  UpdateCollectionRequest,
  CollectionWithFields,
} from "../services/collectionsService";

interface CollectionFormData {
  collection: string;
  note?: string;
  hidden: boolean;
  singleton: boolean;
  accountability: string;
  collapse: string;
  versioning: boolean;
  fields: FieldFormData[];
}

interface FieldFormData {
  field: string;
  interface?: string;
  required: boolean;
  hidden: boolean;
  readonly: boolean;
  width: string;
  note?: string;
}

const Collections: React.FC = () => {
  const [collections, setCollections] = useState<Collection[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingCollection, setEditingCollection] =
    useState<CollectionWithFields | null>(null);
  const [formData, setFormData] = useState<CollectionFormData>({
    collection: "",
    note: "",
    hidden: false,
    singleton: false,
    accountability: "all",
    collapse: "open",
    versioning: false,
    fields: [],
  });

  // Load collections from API
  useEffect(() => {
    loadCollections();
  }, []);

  const loadCollections = async () => {
    try {
      setLoading(true);
      const response = await collectionsService.getCollections();
      setCollections(response.data || []);
    } catch (error) {
      setError("Failed to load collections");
      console.error("Error loading collections:", error);
    } finally {
      setLoading(false);
    }
  };

  const handleOpenDialog = async (collection?: Collection) => {
    if (collection) {
      try {
        // Fetch full collection details with fields
        const collectionDetails = await collectionsService.getCollection(
          collection.collection
        );
        setEditingCollection(collectionDetails);
        setFormData({
          collection: collectionDetails.collection.collection,
          note: collectionDetails.collection.note || "",
          hidden: collectionDetails.collection.hidden,
          singleton: collectionDetails.collection.singleton,
          accountability: collectionDetails.collection.accountability,
          collapse: collectionDetails.collection.collapse,
          versioning: collectionDetails.collection.versioning,
          fields: collectionDetails.fields.map((field) => ({
            field: field.field,
            interface: field.interface,
            required: field.required,
            hidden: field.hidden,
            readonly: field.readonly,
            width: field.width,
            note: field.note,
          })),
        });
      } catch (error) {
        setError("Failed to load collection details");
        console.error("Error loading collection details:", error);
        return;
      }
    } else {
      setEditingCollection(null);
      setFormData({
        collection: "",
        note: "",
        hidden: false,
        singleton: false,
        accountability: "all",
        collapse: "open",
        versioning: false,
        fields: [],
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
  };

  const addField = () => {
    const newField: FieldFormData = {
      field: "",
      interface: "input",
      required: false,
      hidden: false,
      readonly: false,
      width: "full",
      note: "",
    };
    setFormData((prev) => ({
      ...prev,
      fields: [...prev.fields, newField],
    }));
  };

  const updateField = (index: number, field: Partial<FieldFormData>) => {
    setFormData((prev) => ({
      ...prev,
      fields: prev.fields.map((f, i) => (i === index ? { ...f, ...field } : f)),
    }));
  };

  const removeField = (index: number) => {
    setFormData((prev) => ({
      ...prev,
      fields: prev.fields.filter((_, i) => i !== index),
    }));
  };

  const handleSubmit = async () => {
    setError(null);

    // Basic validation
    if (!formData.collection) {
      setError("Please fill in collection name");
      return;
    }

    if (formData.fields.length === 0) {
      setError("At least one field is required");
      return;
    }

    // Validate fields
    for (const field of formData.fields) {
      if (!field.field) {
        setError("All fields must have a name");
        return;
      }
    }

    try {
      if (editingCollection) {
        // Update existing collection
        const updateRequest: UpdateCollectionRequest = {
          note: formData.note,
          hidden: formData.hidden,
          singleton: formData.singleton,
          accountability: formData.accountability,
          collapse: formData.collapse,
          versioning: formData.versioning,
        };

        await collectionsService.updateCollection(
          editingCollection.collection.collection,
          updateRequest
        );
      } else {
        // Create new collection
        const createRequest: CreateCollectionRequest = {
          collection: formData.collection,
          hidden: formData.hidden,
          singleton: formData.singleton,
          accountability: formData.accountability,
          collapse: formData.collapse,
          versioning: formData.versioning,
          fields: formData.fields.map((field) => {
            const fieldData: any = {
              id: "", // Will be generated by backend
              collection: formData.collection,
              field: field.field,
              required: field.required,
              hidden: field.hidden,
              readonly: field.readonly,
              width: field.width,
              special: [],
            };

            // Only include optional fields if they have values
            if (field.interface) fieldData.interface = field.interface;
            if (field.note) fieldData.note = field.note;

            return fieldData;
          }),
        };

        // Add note if it has a value
        if (formData.note) {
          createRequest.note = formData.note;
        }

        await collectionsService.createCollection(createRequest);
      }

      // Reload collections
      await loadCollections();
      handleCloseDialog();
    } catch (err: any) {
      setError(err.response?.data?.error || "Failed to save collection");
    }
  };

  const handleDelete = async (collectionName: string) => {
    if (
      window.confirm(
        "Are you sure you want to delete this collection? This will also delete all data in the collection."
      )
    ) {
      try {
        await collectionsService.deleteCollection(collectionName);
        await loadCollections();
      } catch (error: any) {
        setError(error.response?.data?.error || "Failed to delete collection");
      }
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
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
          <Grid item xs={12} md={6} lg={4} key={collection.collection}>
            <Card
              sx={{ height: "100%", display: "flex", flexDirection: "column" }}
            >
              <CardContent sx={{ flexGrow: 1 }}>
                <Box
                  sx={{ display: "flex", alignItems: "center", gap: 1, mb: 2 }}
                >
                  <StorageIcon color="primary" />
                  <Typography variant="h6" component="div">
                    {collection.collection}
                  </Typography>
                </Box>

                {collection.note && (
                  <Typography
                    variant="body2"
                    color="text.secondary"
                    sx={{ mb: 1 }}
                  >
                    Note: {collection.note}
                  </Typography>
                )}

                <Box sx={{ display: "flex", gap: 1, mb: 2, flexWrap: "wrap" }}>
                  {collection.hidden && (
                    <Chip label="Hidden" size="small" color="warning" />
                  )}
                  {collection.singleton && (
                    <Chip label="Singleton" size="small" color="info" />
                  )}
                  {collection.versioning && (
                    <Chip label="Versioning" size="small" color="secondary" />
                  )}
                </Box>

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
                  onClick={() => handleDelete(collection.collection)}
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
              value={formData.collection}
              onChange={(e) => handleFormChange("collection", e.target.value)}
              required
              fullWidth
              disabled={!!editingCollection}
              helperText={
                editingCollection
                  ? "Collection name cannot be changed"
                  : "The name of the collection (database table)"
              }
            />

            <TextField
              label="Note"
              value={formData.note}
              onChange={(e) => handleFormChange("note", e.target.value)}
              fullWidth
              multiline
              rows={2}
              helperText="Optional description for this collection"
            />

            <Grid container spacing={2}>
              <Grid item xs={6}>
                <FormControl fullWidth size="small">
                  <InputLabel>Accountability</InputLabel>
                  <Select
                    value={formData.accountability}
                    label="Accountability"
                    onChange={(e) =>
                      handleFormChange("accountability", e.target.value)
                    }
                  >
                    <MenuItem value="all">All</MenuItem>
                    <MenuItem value="activity">Activity</MenuItem>
                    <MenuItem value="none">None</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={6}>
                <FormControl fullWidth size="small">
                  <InputLabel>Collapse</InputLabel>
                  <Select
                    value={formData.collapse}
                    label="Collapse"
                    onChange={(e) =>
                      handleFormChange("collapse", e.target.value)
                    }
                  >
                    <MenuItem value="open">Open</MenuItem>
                    <MenuItem value="closed">Closed</MenuItem>
                    <MenuItem value="locked">Locked</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
            </Grid>

            <Grid container spacing={2}>
              <Grid item xs={4}>
                <FormControl fullWidth size="small">
                  <InputLabel>Hidden</InputLabel>
                  <Select
                    value={formData.hidden ? "yes" : "no"}
                    label="Hidden"
                    onChange={(e) =>
                      handleFormChange("hidden", e.target.value === "yes")
                    }
                  >
                    <MenuItem value="no">No</MenuItem>
                    <MenuItem value="yes">Yes</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={4}>
                <FormControl fullWidth size="small">
                  <InputLabel>Singleton</InputLabel>
                  <Select
                    value={formData.singleton ? "yes" : "no"}
                    label="Singleton"
                    onChange={(e) =>
                      handleFormChange("singleton", e.target.value === "yes")
                    }
                  >
                    <MenuItem value="no">No</MenuItem>
                    <MenuItem value="yes">Yes</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
              <Grid item xs={4}>
                <FormControl fullWidth size="small">
                  <InputLabel>Versioning</InputLabel>
                  <Select
                    value={formData.versioning ? "yes" : "no"}
                    label="Versioning"
                    onChange={(e) =>
                      handleFormChange("versioning", e.target.value === "yes")
                    }
                  >
                    <MenuItem value="no">No</MenuItem>
                    <MenuItem value="yes">Yes</MenuItem>
                  </Select>
                </FormControl>
              </Grid>
            </Grid>

            <Box sx={{ mt: 2 }}>
              <Box
                sx={{
                  display: "flex",
                  justifyContent: "space-between",
                  alignItems: "center",
                  mb: 2,
                }}
              >
                <Typography variant="h6">Fields</Typography>
                <Button
                  startIcon={<AddIcon />}
                  onClick={addField}
                  disabled={!!editingCollection}
                >
                  Add Field
                </Button>
              </Box>

              {formData.fields.map((field, index) => (
                <Paper key={index} sx={{ p: 2, mb: 2 }}>
                  <Box
                    sx={{ display: "flex", flexDirection: "column", gap: 2 }}
                  >
                    <Box sx={{ display: "flex", gap: 2, alignItems: "center" }}>
                      <TextField
                        label="Field Name"
                        value={field.field}
                        onChange={(e) =>
                          updateField(index, { field: e.target.value })
                        }
                        required
                        size="small"
                        sx={{ flex: 1 }}
                        disabled={!!editingCollection}
                      />
                      <FormControl size="small" sx={{ minWidth: 120 }}>
                        <InputLabel>Interface</InputLabel>
                        <Select
                          value={field.interface || "input"}
                          label="Interface"
                          onChange={(e) =>
                            updateField(index, {
                              interface: e.target.value,
                            })
                          }
                          disabled={!!editingCollection}
                        >
                          <MenuItem value="input">Input</MenuItem>
                          <MenuItem value="textarea">Textarea</MenuItem>
                          <MenuItem value="select-dropdown">Select</MenuItem>
                          <MenuItem value="datetime">DateTime</MenuItem>
                          <MenuItem value="toggle">Toggle</MenuItem>
                          <MenuItem value="file">File</MenuItem>
                        </Select>
                      </FormControl>
                      <IconButton
                        onClick={() => removeField(index)}
                        color="error"
                        size="small"
                        disabled={!!editingCollection}
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
                          disabled={!!editingCollection}
                        >
                          <MenuItem value="no">No</MenuItem>
                          <MenuItem value="yes">Yes</MenuItem>
                        </Select>
                      </FormControl>

                      <FormControl size="small">
                        <InputLabel>Hidden</InputLabel>
                        <Select
                          value={field.hidden ? "yes" : "no"}
                          label="Hidden"
                          onChange={(e) =>
                            updateField(index, {
                              hidden: e.target.value === "yes",
                            })
                          }
                          disabled={!!editingCollection}
                        >
                          <MenuItem value="no">No</MenuItem>
                          <MenuItem value="yes">Yes</MenuItem>
                        </Select>
                      </FormControl>

                      <FormControl size="small">
                        <InputLabel>Read Only</InputLabel>
                        <Select
                          value={field.readonly ? "yes" : "no"}
                          label="Read Only"
                          onChange={(e) =>
                            updateField(index, {
                              readonly: e.target.value === "yes",
                            })
                          }
                          disabled={!!editingCollection}
                        >
                          <MenuItem value="no">No</MenuItem>
                          <MenuItem value="yes">Yes</MenuItem>
                        </Select>
                      </FormControl>

                      <FormControl size="small" sx={{ minWidth: 100 }}>
                        <InputLabel>Width</InputLabel>
                        <Select
                          value={field.width}
                          label="Width"
                          onChange={(e) =>
                            updateField(index, {
                              width: e.target.value,
                            })
                          }
                          disabled={!!editingCollection}
                        >
                          <MenuItem value="half">Half</MenuItem>
                          <MenuItem value="full">Full</MenuItem>
                        </Select>
                      </FormControl>

                      <TextField
                        label="Note"
                        value={field.note || ""}
                        onChange={(e) =>
                          updateField(index, {
                            note: e.target.value,
                          })
                        }
                        size="small"
                        sx={{ flex: 1 }}
                        disabled={!!editingCollection}
                      />
                    </Box>
                  </Box>
                </Paper>
              ))}

              {editingCollection && (
                <Alert severity="info" sx={{ mt: 2 }}>
                  Field editing is not yet supported. Please create a new
                  collection to modify fields.
                </Alert>
              )}
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
