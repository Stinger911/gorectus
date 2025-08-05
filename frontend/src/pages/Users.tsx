import React, { useState, useEffect } from "react";
import {
  Box,
  Typography,
  Button,
  Paper,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
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
  Pagination,
} from "@mui/material";
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Visibility as ViewIcon,
} from "@mui/icons-material";
import {
  User,
  CreateUserRequest,
  UpdateUserRequest,
  usersService,
} from "../services/usersService";
import { Role, rolesService } from "../services/rolesService";

interface UserFormData {
  email: string;
  first_name: string;
  last_name: string;
  password?: string;
  role_id: string;
  status: string;
  language?: string;
  theme?: string;
  email_notifications?: boolean;
}

const Users: React.FC = () => {
  const [users, setUsers] = useState<User[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | null>(null);
  const [currentPage, setCurrentPage] = useState(1);
  const [totalPages, setTotalPages] = useState(1);
  const [totalUsers, setTotalUsers] = useState(0);
  const [submitting, setSubmitting] = useState(false);
  const [formData, setFormData] = useState<UserFormData>({
    email: "",
    first_name: "",
    last_name: "",
    password: "",
    role_id: "",
    status: "active",
    language: "en-US",
    theme: "auto",
    email_notifications: true,
  });

  // Fetch users and roles on component mount
  useEffect(() => {
    const fetchData = async () => {
      try {
        setLoading(true);
        setError(null);

        // Fetch roles first
        const rolesResponse = await rolesService.getRoles(1, 100); // Get all roles
        setRoles(rolesResponse.data);

        // Then fetch users
        await fetchUsers(1); // Always start from page 1 initially
      } catch (err: any) {
        setError(err.message || "Failed to load data");
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, []); // Remove currentPage dependency

  // Fetch users with pagination
  const fetchUsers = async (page: number) => {
    try {
      const response = await usersService.getUsers(page, 20);
      setUsers(response.data);
      setCurrentPage(page);
      setTotalUsers(response.meta.total);
      setTotalPages(Math.ceil(response.meta.total / response.meta.limit));
    } catch (err: any) {
      setError(err.message || "Failed to fetch users");
    }
  };

  const handlePageChange = (
    event: React.ChangeEvent<unknown>,
    page: number
  ) => {
    fetchUsers(page);
  };

  const handleOpenDialog = (user?: User) => {
    if (user) {
      setEditingUser(user);
      setFormData({
        email: user.email,
        first_name: user.first_name,
        last_name: user.last_name,
        role_id: user.role_id,
        status: user.status,
        language: user.language,
        theme: user.theme,
        email_notifications: user.email_notifications,
      });
    } else {
      setEditingUser(null);
      setFormData({
        email: "",
        first_name: "",
        last_name: "",
        password: "",
        role_id: roles.length > 0 ? roles[0].id : "",
        status: "active",
        language: "en-US",
        theme: "auto",
        email_notifications: true,
      });
    }
    setDialogOpen(true);
  };

  const handleCloseDialog = () => {
    setDialogOpen(false);
    setEditingUser(null);
    setError(null);
    setSubmitting(false);
  };

  const handleFormChange = (
    field: keyof UserFormData,
    value: string | boolean
  ) => {
    setFormData((prev) => ({ ...prev, [field]: value }));
  };

  const handleSubmit = async () => {
    setError(null);
    setSubmitting(true);

    // Basic validation
    if (
      !formData.email ||
      !formData.first_name ||
      !formData.last_name ||
      !formData.role_id
    ) {
      setError("Please fill in all required fields");
      setSubmitting(false);
      return;
    }

    if (!editingUser && !formData.password) {
      setError("Password is required for new users");
      setSubmitting(false);
      return;
    }

    try {
      if (editingUser) {
        // Update existing user
        const updateData: UpdateUserRequest = {
          email: formData.email,
          first_name: formData.first_name,
          last_name: formData.last_name,
          role_id: formData.role_id,
          status: formData.status,
          language: formData.language,
          theme: formData.theme,
          email_notifications: formData.email_notifications,
        };

        if (formData.password) {
          updateData.password = formData.password;
        }

        await usersService.updateUser(editingUser.id, updateData);
      } else {
        // Create new user
        const createData: CreateUserRequest = {
          email: formData.email,
          password: formData.password!,
          first_name: formData.first_name,
          last_name: formData.last_name,
          role_id: formData.role_id,
          status: formData.status,
          language: formData.language,
          theme: formData.theme,
          email_notifications: formData.email_notifications,
        };

        await usersService.createUser(createData);
      }

      // Refresh users list
      await fetchUsers(currentPage);
      handleCloseDialog();
    } catch (err: any) {
      setError(err.message || "Failed to save user");
    } finally {
      setSubmitting(false);
    }
  };

  const handleDelete = async (userId: string) => {
    if (window.confirm("Are you sure you want to delete this user?")) {
      try {
        await usersService.deleteUser(userId);
        // Refresh users list
        await fetchUsers(currentPage);
      } catch (err: any) {
        setError(err.message || "Failed to delete user");
      }
    }
  };

  const formatDate = (dateString: string | null | undefined) => {
    if (!dateString) return "Never";
    return new Date(dateString).toLocaleDateString();
  };

  const getRoleColor = (roleName: string) => {
    switch (roleName.toLowerCase()) {
      case "administrator":
        return "error";
      case "editor":
        return "warning";
      case "moderator":
        return "info";
      default:
        return "default";
    }
  };

  const getStatusColor = (status: string) => {
    return status === "active" ? "success" : "default";
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
      {error && (
        <Alert severity="error" sx={{ mb: 2 }} onClose={() => setError(null)}>
          {error}
        </Alert>
      )}

      <Box
        sx={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          mb: 3,
        }}
      >
        <Typography variant="h4" sx={{ fontWeight: "bold" }}>
          Users ({totalUsers})
        </Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
          disabled={loading}
        >
          Add User
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Name</TableCell>
              <TableCell>Email</TableCell>
              <TableCell>Role</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Created</TableCell>
              <TableCell>Last Access</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {loading ? (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                  <CircularProgress />
                </TableCell>
              </TableRow>
            ) : users.length === 0 ? (
              <TableRow>
                <TableCell colSpan={7} align="center" sx={{ py: 4 }}>
                  <Typography variant="body2" color="text.secondary">
                    No users found
                  </Typography>
                </TableCell>
              </TableRow>
            ) : (
              users.map((user) => (
                <TableRow key={user.id}>
                  <TableCell>
                    <Typography variant="body2" sx={{ fontWeight: "medium" }}>
                      {user.first_name} {user.last_name}
                    </Typography>
                  </TableCell>
                  <TableCell>{user.email}</TableCell>
                  <TableCell>
                    <Chip
                      label={user.role_name}
                      size="small"
                      color={getRoleColor(user.role_name) as any}
                    />
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={user.status}
                      size="small"
                      color={getStatusColor(user.status) as any}
                    />
                  </TableCell>
                  <TableCell>{formatDate(user.created_at)}</TableCell>
                  <TableCell>{formatDate(user.last_access)}</TableCell>
                  <TableCell align="right">
                    <IconButton size="small" title="View">
                      <ViewIcon />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleOpenDialog(user)}
                      title="Edit"
                    >
                      <EditIcon />
                    </IconButton>
                    <IconButton
                      size="small"
                      onClick={() => handleDelete(user.id)}
                      title="Delete"
                      color="error"
                    >
                      <DeleteIcon />
                    </IconButton>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Pagination */}
      {totalPages > 1 && (
        <Box sx={{ display: "flex", justifyContent: "center", mt: 3 }}>
          <Pagination
            count={totalPages}
            page={currentPage}
            onChange={handlePageChange}
            color="primary"
          />
        </Box>
      )}

      <Dialog
        open={dialogOpen}
        onClose={handleCloseDialog}
        maxWidth="sm"
        fullWidth
      >
        <DialogTitle>{editingUser ? "Edit User" : "Add New User"}</DialogTitle>
        <DialogContent>
          {error && (
            <Alert severity="error" sx={{ mb: 2 }}>
              {error}
            </Alert>
          )}
          <Box sx={{ display: "flex", flexDirection: "column", gap: 2, mt: 1 }}>
            <TextField
              label="Email"
              type="email"
              value={formData.email}
              onChange={(e) => handleFormChange("email", e.target.value)}
              required
              fullWidth
              disabled={submitting}
            />
            <Box sx={{ display: "flex", gap: 2 }}>
              <TextField
                label="First Name"
                value={formData.first_name}
                onChange={(e) => handleFormChange("first_name", e.target.value)}
                required
                fullWidth
                disabled={submitting}
              />
              <TextField
                label="Last Name"
                value={formData.last_name}
                onChange={(e) => handleFormChange("last_name", e.target.value)}
                required
                fullWidth
                disabled={submitting}
              />
            </Box>
            <TextField
              label={
                editingUser
                  ? "New Password (leave blank to keep current)"
                  : "Password"
              }
              type="password"
              value={formData.password || ""}
              onChange={(e) => handleFormChange("password", e.target.value)}
              required={!editingUser}
              fullWidth
              disabled={submitting}
            />
            <Box sx={{ display: "flex", gap: 2 }}>
              <FormControl fullWidth disabled={submitting}>
                <InputLabel>Role</InputLabel>
                <Select
                  value={formData.role_id}
                  label="Role"
                  onChange={(e) => handleFormChange("role_id", e.target.value)}
                >
                  {roles.map((role) => (
                    <MenuItem key={role.id} value={role.id}>
                      {role.name}
                    </MenuItem>
                  ))}
                </Select>
              </FormControl>
              <FormControl fullWidth disabled={submitting}>
                <InputLabel>Status</InputLabel>
                <Select
                  value={formData.status}
                  label="Status"
                  onChange={(e) => handleFormChange("status", e.target.value)}
                >
                  <MenuItem value="active">Active</MenuItem>
                  <MenuItem value="inactive">Inactive</MenuItem>
                </Select>
              </FormControl>
            </Box>
            <Box sx={{ display: "flex", gap: 2 }}>
              <FormControl fullWidth disabled={submitting}>
                <InputLabel>Language</InputLabel>
                <Select
                  value={formData.language || "en-US"}
                  label="Language"
                  onChange={(e) => handleFormChange("language", e.target.value)}
                >
                  <MenuItem value="en-US">English (US)</MenuItem>
                  <MenuItem value="en-GB">English (UK)</MenuItem>
                  <MenuItem value="es-ES">Spanish</MenuItem>
                  <MenuItem value="fr-FR">French</MenuItem>
                  <MenuItem value="de-DE">German</MenuItem>
                </Select>
              </FormControl>
              <FormControl fullWidth disabled={submitting}>
                <InputLabel>Theme</InputLabel>
                <Select
                  value={formData.theme || "auto"}
                  label="Theme"
                  onChange={(e) => handleFormChange("theme", e.target.value)}
                >
                  <MenuItem value="auto">Auto</MenuItem>
                  <MenuItem value="light">Light</MenuItem>
                  <MenuItem value="dark">Dark</MenuItem>
                </Select>
              </FormControl>
            </Box>
            <FormControl fullWidth disabled={submitting}>
              <InputLabel>Email Notifications</InputLabel>
              <Select
                value={formData.email_notifications ? "true" : "false"}
                label="Email Notifications"
                onChange={(e) =>
                  handleFormChange(
                    "email_notifications",
                    e.target.value === "true"
                  )
                }
              >
                <MenuItem value="true">Enabled</MenuItem>
                <MenuItem value="false">Disabled</MenuItem>
              </Select>
            </FormControl>
          </Box>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog} disabled={submitting}>
            Cancel
          </Button>
          <Button
            onClick={handleSubmit}
            variant="contained"
            disabled={submitting}
            startIcon={submitting ? <CircularProgress size={16} /> : null}
          >
            {submitting ? "Saving..." : editingUser ? "Update" : "Create"}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};

export default Users;
