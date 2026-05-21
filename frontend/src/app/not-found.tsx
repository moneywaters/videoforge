import Link from 'next/link';
import { Button } from '@/components/ui/button';
import { Icons } from '@/components/icons';

export default function NotFound() {
  return (
    <div className="flex items-center justify-center min-h-[80vh]">
      <div className="text-center space-y-4">
        <div className="flex justify-center">
          <Icons.alertCircle className="h-16 w-16 text-muted-foreground" />
        </div>
        <h1 className="text-2xl font-bold">404 - Page Not Found</h1>
        <p className="text-muted-foreground">The page you are looking for does not exist.</p>
        <Link href="/dashboard">
          <Button>
            <Icons.chevronLeft className="mr-2 h-4 w-4" />
            Back to Dashboard
          </Button>
        </Link>
      </div>
    </div>
  );
}