import { useState, useRef, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';
import { Button } from '@/components/ui/Button';
import { Card, CardBody, CardHeader, CardTitle } from '@/components/ui/Card';
import { Input } from '@/components/ui/Input';
import { Textarea } from '@/components/ui/Textarea';
import type { Brief } from '@/types/index';

interface ChatMessage {
  id: string;
  role: 'user' | 'assistant';
  content: string;
}

const AI_QUESTIONS = [
  "Who is your target audience for this video? (e.g., age range, interests, demographics)",
  "What tone and style would you like the video to have? (e.g., fun and energetic, professional, casual)",
  "What is the main message or call-to-action you want viewers to take?",
  "Are there any specific products, features, or benefits you want highlighted?",
  "Do you have any preference for video length or format? (e.g., 15s, 30s, 60s)",
];

const INITIAL_GREETING = "Hi there! I'm here to help you create a detailed brief. I'll ask you a few questions to gather all the important details. Let's get started!";

type FormData = {
  title: string;
  description: string;
  status: 'draft' | 'published' | 'closed';
  bountyBudget: number;
  submissionLimit: number;
  tags: string[];
  deadline?: string;
};

export function BriefCreate() {
  const navigate = useNavigate();
  const [step, setStep] = useState<'form' | 'interview'>('form');
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
    mutationFn: (data: Omit<Brief, 'id' | 'createdAt' | 'currentSubmissions'>) => api.createBrief(data),
    onSuccess: (newBrief) => {
      navigate(`/briefs/${newBrief.id}`);
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
          { id: (Date.now() + 1).toString(), role: 'assistant', content: AI_QUESTIONS[nextQ] },
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

  if (step === 'interview') {
    const showGenerateButton = userAnswers.length >= AI_QUESTIONS.length && createBriefMutation.isPending;

    return (
      <div className="max-w-2xl mx-auto space-y-6">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900">AI Interview</h1>
          <p className="text-gray-500">Answer a few questions to create your brief</p>
        </div>

        <Card className="h-[500px] flex flex-col">
          <CardHeader className="border-b">
            <CardTitle>{formData.title}</CardTitle>
          </CardHeader>
          <CardBody className="flex-1 overflow-y-auto space-y-4">
            {chatMessages.map((msg) => (
              <div
                key={msg.id}
                className={`flex ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                <div
                  className={`max-w-[80%] rounded-lg px-4 py-2 ${
                    msg.role === 'user'
                      ? 'bg-brand-600 text-white'
                      : 'bg-gray-100 text-gray-900'
                  }`}
                >
                  <p className="text-sm">{msg.content}</p>
                </div>
              </div>
            ))}
            {showGenerateButton && (
              <div className="text-center text-gray-500 text-sm">
                Generating brief...
              </div>
            )}
            <div ref={chatEndRef} />
          </CardBody>
          <div className="p-4 border-t">
            {userAnswers.length >= AI_QUESTIONS.length ? (
              <Button
                className="w-full"
                onClick={() => generateBriefFromAnswers(userAnswers)}
                disabled={createBriefMutation.isPending}
              >
                {createBriefMutation.isPending ? 'Creating Brief...' : 'Generate Brief'}
              </Button>
            ) : (
              <div className="flex gap-2">
                <Input
                  value={userAnswer}
                  onChange={(e) => setUserAnswer(e.target.value)}
                  onKeyDown={handleKeyDown}
                  placeholder="Type your answer..."
                  disabled={createBriefMutation.isPending}
                />
                <Button onClick={handleSendMessage} disabled={!userAnswer.trim() || createBriefMutation.isPending}>
                  Send
                </Button>
              </div>
            )}
          </div>
        </Card>
      </div>
    );
  }

  return (
    <div className="max-w-2xl mx-auto space-y-6">
      <div className="text-center">
        <h1 className="text-2xl font-bold text-gray-900">Create New Brief</h1>
        <p className="text-gray-500">Fill in the basic details, then our AI will help you refine</p>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Basic Information</CardTitle>
        </CardHeader>
        <CardBody className="space-y-4">
          <Input
            label="Brief Title"
            value={formData.title}
            onChange={(e) => setFormData({ ...formData, title: e.target.value })}
            placeholder="e.g., Summer Sale TikTok Promo"
          />

          <Textarea
            label="Description (Optional)"
            value={formData.description}
            onChange={(e) => setFormData({ ...formData, description: e.target.value })}
            placeholder="Any initial details you'd like to include..."
          />

          <div className="grid grid-cols-2 gap-4">
            <Input
              label="Bounty Budget ($)"
              type="number"
              value={formData.bountyBudget}
              onChange={(e) => setFormData({ ...formData, bountyBudget: parseInt(e.target.value) || 0 })}
            />
            <Input
              label="Submission Limit"
              type="number"
              value={formData.submissionLimit}
              onChange={(e) => setFormData({ ...formData, submissionLimit: parseInt(e.target.value) || 0 })}
            />
          </div>

          <Input
            label="Deadline (Optional)"
            type="datetime-local"
            value={formData.deadline || ''}
            onChange={(e) => setFormData({ ...formData, deadline: e.target.value })}
          />

          <Input
            label="Tags (comma-separated)"
            value={formData.tags.join(', ')}
            onChange={(e) => setFormData({ ...formData, tags: e.target.value.split(',').map((t) => t.trim()).filter(Boolean) })}
            placeholder="e.g., fashion, summer, sale"
          />
        </CardBody>
      </Card>

      <Button
        className="w-full"
        size="lg"
        onClick={handleStartInterview}
        disabled={!formData.title.trim()}
      >
        Start AI Interview
      </Button>
    </div>
  );
}