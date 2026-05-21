import { cn } from "@/lib/utils";
import { ArrowUp, ArrowDown } from "lucide-react";
import { Card, CardBody } from "./Card";

interface MetricCardProps {
  title: string;
  value: string | number;
  change?: number;
  trend?: 'up' | 'down';
  className?: string;
}

export function MetricCard({ title, value, change, trend, className }: MetricCardProps) {
  return (
    <Card className={className}>
      <CardBody>
        <p className="text-sm font-medium text-gray-500">{title}</p>
        <div className="mt-2 flex items-baseline gap-2">
          <p className="text-3xl font-semibold text-gray-900">{value}</p>
          {change !== undefined && trend && (
            <span
              className={cn(
                "inline-flex items-center text-sm font-medium",
                trend === "up" ? "text-emerald-600" : "text-rose-600"
              )}
            >
              {trend === "up" ? (
                <ArrowUp className="mr-0.5 h-3 w-3" />
              ) : (
                <ArrowDown className="mr-0.5 h-3 w-3" />
              )}
              {Math.abs(change)}%
            </span>
          )}
        </div>
      </CardBody>
    </Card>
  );
}