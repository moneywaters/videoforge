'use client';

import { useState, useRef, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import { useMutation } from '@tanstack/react-query';
import { toast } from 'sonner';
import { api } from '@/lib/api';
import { Icons } from '@/components/icons';
import { Button } from '@/components/ui/button';
import { isClient, hasHydrated, useAuthStore } from '@/stores/auth-store';
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from '@/components/ui/card';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Textarea } from '@/components/ui/textarea';
import type { Brief } from '@/types';

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

const AI_QUESTIONS = [
  'Who is your target audience for this video? (e.g., age range, interests, demographics)',
  'What tone and style would you like the video to have? (e.g., fun and energetic, professional, casual)',
  'What is the main message or call-to-action you want viewers to take?',
  'Are there any specific products, features, or benefits you want highlighted?',
  'Do you have any preference for video length or format? (e.g., 15s, 30s, 60s)',
];

const INITIAL_GREETING =
  "Hi there! I'm here to help you create a detailed brief. I'll ask you a few questions to gather all the important details. Let's get started!";

type FormData = {
  title: string;
  description: string;
  status: 'draft' | 'published' | 'closed';
  bountyBudget: number;
  submissionLimit: number;
  tags: string[];
  deadline?: string;
};

export default function NewBriefPage() {
  const router = useRouter();
  const [step, setStep] = useState<'form' | 'interview'>('form');
  const [accessChecked, setAccessChecked] = useState(false);

  // Restrict brief creation to client role only
  useEffect(() => {
    if (hasHydrated()) {
      // Already hydrated, check immediately
      if (!isClient()) {
        toast.error('Only clients can create briefs');
        router.push('/dashboard/briefs');
      } else {
        setAccessChecked(true);
      }
    } else {
      // Wait for hydration
      const unsubscribe = useAuthStore.subscribe((state) => {
        if (state._hasHydrated) {
          if (!isClient()) {
            toast.error('Only clients can create briefs');
            router.push('/dashboard/briefs');
          } else {
            setAccessChecked(true);
          }
          unsubscribe();
        }
      });
      return () => unsubscribe();
    }
  }, [router]);
  const [formData, setFormData] = useState<FormData>({
    title: '',
    description: '',
    status: 'draft',
    bountyBudget: 500,
    submissionLimit: 5,
    tags: [],
    deadline: '',
  });
  const [chatMessages, setChatMessages] = useState<ChatMessage[]>([
    { id: '1', role: 'assistant', content: INITIAL_GREETING },
  ]);
  const [currentQuestion, setCurrentQuestion] = useState(0);
  const [userAnswer, setUserAnswer] = useState('');
  const [userAnswers, setUserAnswers] = useState<string[]>([]);
  const chatEndRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    chatEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [chatMessages]);

  useEffect(() => {
    if (step === 'interview' && chatMessages.length === 1) {
      setTimeout(() => {
        setChatMessages((prev) => [
          ...prev,
          { id: '2', role: 'assistant', content: AI_QUESTIONS[0] },
        ]);
      }, 500);
    }
  }, [step, chatMessages.length]);

  const createBriefMutation = useMutation({
    mutationFn: (data: Omit<Brief, 'id' | 'createdAt' | 'currentSubmissions'>) =>
      api.createBrief(data),
    onSuccess: (newBrief) => {
      toast.success('Brief created successfully');
      router.push(`/dashboard/briefs/${newBrief.id}`);
    },
    onError: (error: Error) => {
      toast.error(`Failed to create brief: ${error.message}`);
    },
  });

  const handleStartInterview = () => {
    if (!formData.title.trim()) {
      return;
    }
    setStep('interview');
  };

  const handleSendMessage = () => {
    if (!userAnswer.trim()) return;

    const userMsg: ChatMessage = {
      id: Date.now().toString(),
      role: 'user',
      content: userAnswer,
    };
    setChatMessages((prev) => [...prev, userMsg]);
    setUserAnswers((prev) => [...prev, userAnswer]);
    setUserAnswer('');

    setTimeout(() => {
      const nextQ = currentQuestion + 1;
      if (nextQ < AI_QUESTIONS.length) {
        setCurrentQuestion(nextQ);
        setChatMessages((prev) => [
          ...prev,
          {
            id: (Date.now() + 1).toString(),
            role: 'assistant',
            content: AI_QUESTIONS[nextQ],
          },
        ]);
      } else {
        generateBriefFromAnswers([...userAnswers, userAnswer]);
      }
    }, 800);
  };

  const generateBriefFromAnswers = (answers: string[]) => {
    const targetAudience = answers[0] || '';
    const toneStyle = answers[1] || '';
    const cta = answers[2] || '';
    const features = answers[3] || '';
    const lengthPref = answers[4] || '';

    const generatedDescription = `Target Audience: ${targetAudience}

Tone & Style: ${toneStyle}

Call-to-Action: ${cta}

Key Features to Highlight: ${features}

Format Preference: ${lengthPref}`;

    const descriptionWithContext = formData.description
      ? `${formData.description}\n\n---\n\nAI Interview Details:\n${generatedDescription}`
      : generatedDescription;

    const newFormData = {
      ...formData,
      description: descriptionWithContext,
      status: 'published' as const,
      clientId: 'usr-client-001',
      clientName: 'John Smith',
    };

    createBriefMutation.mutate(newFormData);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSendMessage();
    }
  };

  if (!accessChecked) {
    return (
      <div className='flex items-center justify-center min-h-[400px]'>
        <Icons.spinner className='h-8 w-8 animate-spin text-muted-foreground' />
      </div>
    );
  }

  if (step === 'interview') {
    const allQuestionsAnswered = userAnswers.length >= AI_QUESTIONS.length;

    return (
      <div className='max-w-2xl mx-auto space-y-6'>
        <div className='text-center'>
          <h1 className='text-2xl font-bold text-foreground'>AI Interview</h1>
          <p className='text-muted-foreground'>
            Answer a few questions to create your brief
          </p>
        </div>

        <Card className='h-[500px] flex flex-col'>
          <CardHeader className='border-b shrink-0'>
            <CardTitle>{formData.title}</CardTitle>
          </CardHeader>
          <CardContent className='flex-1 overflow-y-auto space-y-4'>
            {chatMessages.map((msg) => (
              <div
                key={msg.id}
                className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`max-w-[80%] rounded-lg px-4 py-2 ${
                    msg.role === 'user'
                      ? 'bg-primary text-primary-foreground'
                      : 'bg-muted text-foreground'
                  }`}
                >
                  <p className='text-sm'>{msg.content}</p>
                </div>
              </div>
            ))}
            <div ref={chatEndRef} />
          </CardContent>
          <div className='p-4 border-t shrink-0'>
            {allQuestionsAnswered ? (
              <Button
                className='w-full'
                onClick={() => generateBriefFromAnswers(userAnswers)}
                disabled={createBriefMutation.isPending}
              >
                {createBriefMutation.isPending ? (
                  <>
                    <Icons.spinner className='mr-2 h-4 w-4 animate-spin' />
                    Creating Brief...
                  </>
                ) : (
                  'Generate Brief'
                )}
              </Button>
            ) : (
              <div className='flex gap-2'>
                <Input
                  value={userAnswer}
                  onChange={(e) => setUserAnswer(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder='Type your answer...'
                  disabled={createBriefMutation.isPending}
                />
                <Button
                  onClick={handleSendMessage}
                  disabled={!userAnswer.trim() || createBriefMutation.isPending}
                >
                  <Icons.send className='h-4 w-4' />
                </Button>
              </div>
            )}
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className='max-w-2xl mx-auto space-y-6'>
      <div className='text-center'>
        <h1 className='text-2xl font-bold text-foreground'>Create New Brief</h1>
        <p className='text-muted-foreground'>
          Fill in the basic details, then our AI will help you refine
        </p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Basic Information</CardTitle>
          <CardDescription>
            Enter the core details for your brief
          </CardDescription>
        </CardHeader>
        <CardContent className='space-y-4'>
          <div className='space-y-2'>
            <Label htmlFor='title'>Brief Title</Label>
            <Input
              id='title'
              value={formData.title}
              onChange={(e) =>
                setFormData({ ...formData, title: e.target.value })
              }
              placeholder='e.g., Summer Sale TikTok Promo'
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='description'>Description (Optional)</Label>
            <Textarea
              id='description'
              value={formData.description}
              onChange={(e) =>
                setFormData({ ...formData, description: e.target.value })
              }
              placeholder='Any initial details you would like to include...'
            />
          </div>

          <div className='grid grid-cols-2 gap-4'>
            <div className='space-y-2'>
              <Label htmlFor='bountyBudget'>Bounty Budget ($)</Label>
              <Input
                id='bountyBudget'
                type='number'
                value={formData.bountyBudget}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    bountyBudget: parseInt(e.target.value) || 0,
                  })
                }
              />
            </div>
            <div className='space-y-2'>
              <Label htmlFor='submissionLimit'>Submission Limit</Label>
              <Input
                id='submissionLimit'
                type='number'
                value={formData.submissionLimit}
                onChange={(e) =>
                  setFormData({
                    ...formData,
                    submissionLimit: parseInt(e.target.value) || 0,
                  })
                }
              />
            </div>
          </div>

          <div className='space-y-2'>
            <Label htmlFor='deadline'>Deadline (Optional)</Label>
            <Input
              id='deadline'
              type='datetime-local'
              value={formData.deadline || ''}
              onChange={(e) =>
                setFormData({ ...formData, deadline: e.target.value })
              }
            />
          </div>

          <div className='space-y-2'>
            <Label htmlFor='tags'>Tags (comma-separated)</Label>
            <Input
              id='tags'
              value={formData.tags.join(', ')}
              onChange={(e) =>
                setFormData({
                  ...formData,
                  tags: e.target.value
                    .split(',')
                    .map((t) => t.trim())
                    .filter(Boolean),
                })
              }
              placeholder='e.g., fashion, summer, sale'
            />
          </div>
        </CardContent>
      </Card>

      <Button
        className='w-full'
        size='lg'
        onClick={handleStartInterview}
        disabled={!formData.title.trim()}
      >
        <Icons.sparkles className='mr-2 h-4 w-4' />
        Start AI Interview
      </Button>
    </div>
  );
}