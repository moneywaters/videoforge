import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Loader2 } from 'lucide-react';
import { useAuthStore } from '@/stores/authStore';
import type { User } from '@/types/index';

/**
 * Decode a JWT token and extract the payload
 */
function decodeJWT(token: string): Record<string, unknown> {
  try {
    const base64Url = token.split('.')[1];
    const base64 = base64Url.replace(/-/g, '+').replace(/_/g, '/');
    const jsonPayload = decodeURIComponent(
      atob(base64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join('')
    );
    return JSON.parse(jsonPayload);
  } catch {
    return {};
  }
}

export default function GoogleCallback() {
  const navigate = useNavigate();
  const [error, setError] = useState<string | null>(null);
  const setLoading = useAuthStore((state) => state.setLoading);
  const setUser = useAuthStore((state) => state.setUser);

  useEffect(() => {
    // Read token from URL query parameter
    const params = new URLSearchParams(window.location.search);
    const token = params.get('token');

    if (!token) {
      setError('No token provided. Please try logging in again.');
      return;
    }

    // Decode the JWT to extract user claims
    const payload = decodeJWT(token);

    if (!payload.sub) {
      setError('Invalid token. Please try logging in again.');
      return;
    }

    // Extract user data from JWT payload
    const user: User = {
      id: payload.sub as string,
      email: (payload.email as string) || '',
      name: (payload.name as string) || '',
      role: (payload.role as User['role']) || 'client',
      avatar: payload.picture ? (payload.picture as string) : undefined,
      createdAt: new Date().toISOString(),
      onboardingComplete: false, // Google users need to complete onboarding
    };

    // Store token in localStorage
    localStorage.setItem('token', token);

    // Set the user in the auth store
    setUser(user);
    setLoading(false);

    // Navigate to home
    navigate('/');
  }, [navigate, setLoading, setUser]);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
        <div className="bg-white p-8 rounded-lg shadow-md max-w-md text-center">
          <div className="text-rose-600 text-lg font-medium mb-4">Authentication Error</div>
          <p className="text-gray-600 mb-4">{error}</p>
          <a
            href="/login"
            className="inline-block bg-brand-600 text-white px-4 py-2 rounded-md hover:bg-brand-700 transition-colors"
          >
            Back to Login
          </a>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <div className="text-center">
        <Loader2 className="w-8 h-8 animate-spin text-brand-600 mx-auto mb-4" />
        <p className="text-gray-600">Completing authentication...</p>
      </div>
    </div>
  );
}