import { useState, useRef, useEffect } from 'react';
import { useMutation } from '@tanstack/react-query';
import { api } from '@/lib/api';
import type { ChatMessage } from '@/types/index';
import { Card, CardHeader, CardTitle } from '@/components/ui/Card';
import { Button } from '@/components/ui/Button';
import { Input } from '@/components/ui/Input';
import { Send, User, Bot, Headphones } from 'lucide-react';

const WELCOME_MESSAGE: ChatMessage = {
  id: 'welcome',
  role: 'assistant',
  content: "Hi! I'm your VideoForge assistant. How can I help you today?",
  timestamp: new Date().toISOString(),
};

export default function Chat() {
  const [messages, setMessages] = useState<ChatMessage[]>([WELCOME_MESSAGE]);
  const [inputValue, setInputValue] = useState('');
  const [isTyping, setIsTyping] = useState(false);
  const messagesEndRef = useRef<HTMLDivElement>(null);

  const { mutate: sendMessage } = useMutation({
    mutationFn: async (content: string) => {
      const userMessage: ChatMessage = {
        id: `msg-${Date.now()}`,
        role: 'user',
        content,
        timestamp: new Date().toISOString(),
      };

      setMessages((prev) => [...prev, userMessage]);
      setInputValue('');
      setIsTyping(true);

      const response = await api.sendChatMessage(content);
      return response;
    },
    onSuccess: (aiResponse) => {
      setIsTyping(false);
      setMessages((prev) => [
        ...prev,
        {
          ...aiResponse,
          content: getAIResponse(aiResponse.content),
        },
      ]);
    },
    onError: () => {
      setIsTyping(false);
      setMessages((prev) => [
        ...prev,
        {
          id: `msg-${Date.now()}`,
          role: 'assistant',
          content: "I'm sorry, I encountered an error. Please try again.",
          timestamp: new Date().toISOString(),
        },
      ]);
    },
  });

  const getAIResponse = (input: string): string => {
    const lowerInput = input.toLowerCase();

    if (lowerInput.includes('payout') || lowerInput.includes('payment')) {
      return "For payout questions, you can check your Earnings page to see your balance and pending payments. Payouts are processed within 5-7 business days. Is there something specific about your earnings you'd like to know?";
    }
    if (lowerInput.includes('video') || lowerInput.includes('submission')) {
      return "To submit a video, go to the Videos page and select the brief you want to work on. Make sure your video meets the requirements specified in the brief. Need help with video submissions?";
    }
    if (lowerInput.includes('brief') || lowerInput.includes('project')) {
      return "You can find available briefs on the Briefs page. Browse through published briefs and submit your proposal. The client will review your submission and provide feedback. Want help finding the right brief?";
    }
    if (lowerInput.includes('campaign') || lowerInput.includes('ad')) {
      return "Campaign management is available for Ad Specialists. You can launch and manage ad campaigns from the Campaigns page. Need help with campaign setup or optimization?";
    }
    if (lowerInput.includes('hello') || lowerInput.includes('hi') || lowerInput.includes('hey')) {
      return "Hello! I'm here to help you with any questions about VideoForge. What would you like to know?";
    }
    if (lowerInput.includes('thank')) {
      return "You're welcome! Is there anything else I can help you with?";
    }

    return "I understand you need help with that. Let me point you in the right direction - you can navigate to the relevant page from the sidebar menu. If you need more specific assistance, I can connect you with a human support agent.";
  };

  const scrollToBottom = () => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  };

  useEffect(() => {
    scrollToBottom();
  }, [messages, isTyping]);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (inputValue.trim() && !isTyping) {
      sendMessage(inputValue.trim());
    }
  };

  const handleKeyPress = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      handleSubmit(e);
    }
  };

  const handleEscalate = () => {
    const escalateMessage: ChatMessage = {
      id: `msg-${Date.now()}`,
      role: 'assistant',
      content:
        "Your request has been escalated to our human support team. They'll get back to you within 24 hours. Is there anything else I can help you with in the meantime?",
      timestamp: new Date().toISOString(),
    };
    setMessages((prev) => [...prev, escalateMessage]);
  };

  const formatTimestamp = (dateStr: string) => {
    return new Date(dateStr).toLocaleTimeString('en-US', {
      hour: 'numeric',
      minute: '2-digit',
    });
  };

  return (
    <div className="p-8 h-[calc(100vh-2rem)] flex flex-col">
      <div className="mb-6">
        <h1 className="text-3xl font-bold text-gray-900">AI Support</h1>
        <p className="text-gray-500 mt-2">
          Get help from our AI assistant or escalate to human support
        </p>
      </div>

      <Card className="flex-1 flex flex-col overflow-hidden">
        <CardHeader className="flex-shrink-0 border-b border-gray-200">
          <CardTitle className="flex items-center gap-2">
            <Bot className="h-5 w-5 text-brand-600" />
            VideoForge Assistant
          </CardTitle>
        </CardHeader>

        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {messages.map((message) => (
            <div
              key={message.id}
              className={`flex ${
                message.role === 'user' ? 'justify-end' : 'justify-start'
              }`}
            >
              <div
                className={`max-w-[75%] rounded-lg px-4 py-2 ${
                  message.role === 'user'
                    ? 'bg-brand-600 text-white'
                    : 'bg-white border border-gray-200 text-gray-900'
                }`}
              >
                <div className="flex items-start gap-2">
                  {message.role === 'user' ? (
                    <User className="h-5 w-5 mt-0.5 flex-shrink-0" />
                  ) : (
                    <Bot className="h-5 w-5 mt-0.5 flex-shrink-0 text-brand-600" />
                  )}
                  <p className="text-sm whitespace-pre-wrap">{message.content}</p>
                </div>
                <p
                  className={`text-xs mt-2 ${
                    message.role === 'user'
                      ? 'text-brand-100'
                      : 'text-gray-400'
                  }`}
                >
                  {formatTimestamp(message.timestamp)}
                </p>
              </div>
            </div>
          ))}

          {isTyping && (
            <div className="flex justify-start">
              <div className="bg-white border border-gray-200 rounded-lg px-4 py-3">
                <div className="flex items-center gap-2">
                  <Bot className="h-5 w-5 text-brand-600" />
                  <div className="flex items-center gap-1">
                    <span className="h-2 w-2 bg-gray-400 rounded-full animate-bounce [animation-delay:-0.3s]"></span>
                    <span className="h-2 w-2 bg-gray-400 rounded-full animate-bounce [animation-delay:-0.15s]"></span>
                    <span className="h-2 w-2 bg-gray-400 rounded-full animate-bounce"></span>
                  </div>
                </div>
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        <div className="flex-shrink-0 border-t border-gray-200 p-4 space-y-3">
          <form onSubmit={handleSubmit} className="flex gap-2">
            <Input
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              onKeyPress={handleKeyPress}
              placeholder="Type your message..."
              disabled={isTyping}
              className="flex-1"
            />
            <Button type="submit" variant="primary" disabled={isTyping || !inputValue.trim()}>
              <Send className="h-4 w-4" />
            </Button>
          </form>

          <div className="flex justify-center">
            <Button
              type="button"
              variant="ghost"
              onClick={handleEscalate}
              className="text-gray-500"
            >
              <Headphones className="h-4 w-4 mr-2" />
              Escalate to Human
            </Button>
          </div>
        </div>
      </Card>
    </div>
  );
}