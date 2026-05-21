import { useState, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { Button } from '@/components/ui/Button';
import { Card, CardBody, CardHeader, CardTitle, CardDescription } from '@/components/ui/Card';
import { useAuth } from '@/hooks/useAuth';
import { api } from '@/lib/api';
import type { OnboardingStep } from '@/types/index';

export default function Onboarding() {
  const navigate = useNavigate();
  const { user } = useAuth();
  const [steps, setSteps] = useState<OnboardingStep[]>([]);
  const [currentStep, setCurrentStep] = useState(0);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    const loadSteps = async () => {
      try {
        const data = await api.getOnboardingSteps();
        setSteps(data);
      } catch (err) {
        console.error('Failed to load onboarding steps', err);
      } finally {
        setIsLoading(false);
      }
    };
    loadSteps();
  }, []);

  const handleComplete = async () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep((prev) => prev + 1);
    } else {
      // Mark onboarding complete and navigate to dashboard
      if (user) {
        user.onboardingComplete = true;
      }
      navigate('/');
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep((prev) => prev - 1);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-gray-100 flex items-center justify-center">
        <p className="text-gray-500">Loading...</p>
      </div>
    );
  }

  const step = steps[currentStep];

  return (
    <div className="min-h-screen bg-gray-100 flex items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>Welcome to VideoForge</CardTitle>
          <CardDescription>Let's get you set up</CardDescription>
        </CardHeader>
        <CardBody>
          {/* Step Indicator */}
          <div className="mb-8">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-gray-500">
                Step {currentStep + 1} of {steps.length}
              </span>
              <span className="text-sm text-gray-500">
                {Math.round(((currentStep + 1) / steps.length) * 100)}%
              </span>
            </div>
            <div className="h-2 bg-gray-200 rounded-full overflow-hidden">
              <div
                className="h-full bg-brand-600 transition-all duration-300"
                style={{ width: `${((currentStep + 1) / steps.length) * 100}%` }}
              />
            </div>
          </div>

          {/* Step Content */}
          {step && (
            <div className="mb-8">
              <h2 className="text-xl font-semibold text-gray-900 mb-2">
                {step.title}
              </h2>
              <p className="text-gray-600">{step.description}</p>
            </div>
          )}

          {/* Navigation Buttons */}
          <div className="flex justify-between gap-4">
            <Button
              variant="secondary"
              onClick={handleBack}
              disabled={currentStep === 0}
            >
              Back
            </Button>
            <Button onClick={handleComplete}>
              {currentStep === steps.length - 1 ? 'Complete' : 'Next'}
            </Button>
          </div>
        </CardBody>
      </Card>
    </div>
  );
}