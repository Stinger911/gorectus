import React from 'react';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter as Router } from 'react-router-dom';
import { AuthProvider } from '../contexts/AuthContext';
import Login from './Login';
import { authService } from '../services/authService';

// Mock the authService
jest.mock('../services/authService', () => ({
  authService: {
    login: jest.fn(),
    getStoredUser: jest.fn(() => null),
    isAuthenticated: jest.fn(() => false),
  },
}));

const mockLogin = authService.login as jest.Mock;

const renderLoginComponent = () => {
  return render(
    <Router>
      <AuthProvider>
        <Login />
      </AuthProvider>
    </Router>
  );
};

describe('Login Page', () => {
  beforeEach(() => {
    mockLogin.mockClear();
  });

  test('renders login form correctly', () => {
    renderLoginComponent();
    expect(screen.getByLabelText(/Email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Password/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Sign In/i })).toBeInTheDocument();
  });

  // test('shows validation errors for empty fields', async () => {
  //   renderLoginComponent();
  //   fireEvent.click(screen.getByRole('button', { name: /Sign In/i }));

  //   await waitFor(() => {
  //     expect(screen.getByText('Email is required')).toBeInTheDocument();
  //     expect(screen.getByText('Password is required')).toBeInTheDocument();
  //   });

  //   expect(mockLogin).not.toHaveBeenCalled();
  // });

  // test('shows validation error for invalid email', async () => {
  //   renderLoginComponent();
  //   fireEvent.input(screen.getByLabelText(/Email/i), {
  //     target: { value: 'invalid-email' },
  //   });
  //   fireEvent.input(screen.getByLabelText(/Password/i), {
  //     target: { value: 'password123' },
  //   });
  //   fireEvent.click(screen.getByRole('button', { name: /Sign In/i }));

  //   await waitFor(() => {
  //     expect(screen.getByText('Invalid email address')).toBeInTheDocument();
  //   });
  //   expect(mockLogin).not.toHaveBeenCalled();
  // });

  test('calls login service with correct credentials on successful submission', async () => {
    mockLogin.mockResolvedValue({
        token: 'fake-token',
        user: {
            id: '1',
            email: 'admin@gorectus.local',
            role_name: 'Administrator',
        }
    });

    renderLoginComponent();

    fireEvent.input(screen.getByLabelText(/Email/i), {
      target: { value: 'admin@gorectus.local' },
    });
    fireEvent.input(screen.getByLabelText(/Password/i), {
      target: { value: 'password123' },
    });

    fireEvent.click(screen.getByRole('button', { name: /Sign In/i }));

    await waitFor(() => {
      expect(mockLogin).toHaveBeenCalledWith('admin@gorectus.local', 'password123');
      expect(mockLogin).toHaveBeenCalledTimes(1);
    });
  });

  test('displays error message on login failure', async () => {
    const errorMessage = 'Invalid credentials';
    mockLogin.mockRejectedValue(new Error(errorMessage));

    renderLoginComponent();

    fireEvent.input(screen.getByLabelText(/Email/i), {
      target: { value: 'wrong@example.com' },
    });
    fireEvent.input(screen.getByLabelText(/Password/i), {
      target: { value: 'wrongpassword' },
    });

    fireEvent.click(screen.getByRole('button', { name: /Sign In/i }));

    await waitFor(() => {
        expect(mockLogin).toHaveBeenCalledTimes(1);
    });

    await waitFor(() => {
      expect(screen.getByRole('alert')).toHaveTextContent(errorMessage);
    });
  });
});
