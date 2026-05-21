'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Icons } from '@/components/icons';

// Local onboarding steps since API returns empty
const DEFAULT_STEPS = [
  { id: '1', title: 'Welcome to VideoForge', description: 'This guided setup will help you get started with your new account. Let\'s go through a few quick steps together.', completed: false },
  { id: '2', title: 'Set Up Your Profile', description: 'Complete your profile so other users can recognize you and know how to reach out.', completed: false },
  { id: '3', title: 'Choose Your Role', description: 'Select whether you want to create video projects or edit videos.', completed: false },
  { id: '4', title: 'You\'re All Set!', description: 'You\'re ready to start using VideoForge. Check out your dashboard to begin.', completed: false },
];

export default function OnboardingPage() {
  const router = useRouter();
  const [currentStep, setCurrentStep] = useState(0);
  const [isLoading, setIsLoading] = useState(true);
  const [steps] = useState(DEFAULT_STEPS);

  useEffect(() => {
    // Simulate loading
    const timer = setTimeout(() => {
      setIsLoading(false);
    }, 500);
    return () => clearTimeout(timer);
  }, []);

  const handleComplete = async () => {
    if (currentStep < steps.length - 1) {
      setCurrentStep((prev) => prev + 1);
    } else {
      // Mark onboarding complete and navigate to dashboard
      router.push('/dashboard');
    }
  };

  const handleBack = () => {
    if (currentStep > 0) {
      setCurrentStep((prev) => prev - 1);
    }
  };

  if (isLoading) {
    return (
      <div className="min-h-screen bg-muted flex items-center justify-center">
        <div className="flex items-center gap-2">
          <Icons.spinner className="h-4 w-4 animate-spin" />
          <p className="text-muted-foreground">Loading...</p>
        </div>
      </div>
    );
  }

  const step = steps[currentStep];

  return (
    <div className="min-h-screen bg-muted flex items-center justify-center p-4">
      <Card className="w-full max-w-lg">
        <CardHeader>
          <CardTitle>Welcome to VideoForge</CardTitle>
          <CardDescription>Let's get you set up</CardDescription>
        </CardHeader>
        <CardContent>
          {/* Step Indicator */}
          <div className="mb-8">
            <div className="flex items-center justify-between mb-2">
              <span className="text-sm text-muted-foreground">
                Step {currentStep + 1} of {steps.length}
              </span>
              <span className="text-sm text-muted-foreground">
                {Math.round(((currentStep + 1) / steps.length) * 100)}%
              </span>
            </div>
            <div className="h-2 bg-muted rounded-full overflow-hidden">
              <div
                className="h-full bg-primary transition-all duration-300"
                style={{ width: `${((currentStep + 1) / steps.length) * 100}%` }}
              />
            </div>
          </div>

          {/* Step Content */}
          {step && (
            <div className="mb-8">
              <h2 className="text-xl font-semibold mb-2">
                {step.title}
              </h2>
              <p className="text-muted-foreground">{step.description}</p>
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
        </CardContent>
      </Card>
    </div>
  );
}