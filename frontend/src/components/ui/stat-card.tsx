import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Link } from "react-router-dom";
import { TrendingUp } from "lucide-react";

interface StatCardProps {
  title: string;
  value: string | number;
  icon: React.ElementType;
  color?: "default" | "success" | "warning" | "danger" | "info" | "purple";
  subtitle?: string;
  description?: string;
  trend?: { value: number; isPositive: boolean };
  to?: string;
  loading?: boolean;
}

export function StatCard({
  title,
  value,
  icon: Icon,
  color = "default",
  subtitle,
  description,
  trend,
  to,
  loading = false,
}: StatCardProps) {
  const colorClasses = {
    default: "text-muted-foreground",
    success: "text-green-600",
    warning: "text-yellow-600",
    danger: "text-red-600",
    info: "text-blue-600",
    purple: "text-purple-600",
  };
  const bgClasses = {
    default: "bg-muted/20",
    success: "bg-green-100 dark:bg-green-900/20",
    warning: "bg-yellow-100 dark:bg-yellow-900/20",
    danger: "bg-red-100 dark:bg-red-900/20",
    info: "bg-blue-100 dark:bg-blue-900/20",
    purple: "bg-purple-100 dark:bg-purple-900/20",
  };
  
  const content = (
    <Card className="rounded-xl shadow-lg hover:shadow-2xl transition-all duration-300 bg-gradient-to-br from-white to-gray-50 dark:from-gray-900 dark:to-gray-800 border-0">
      <CardHeader className="flex flex-row items-center justify-between pb-3">
        <div className={`p-3 rounded-xl ${bgClasses[color]}`}>
          <Icon className={`h-6 w-6 ${colorClasses[color]}`} />
        </div>
        <CardTitle className="text-lg font-semibold text-gray-900 dark:text-gray-100">
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent>
        <div className="text-3xl font-bold mb-1 text-gray-900 dark:text-gray-100">
          {loading ? "..." : value}
        </div>
        {subtitle && (
          <div className="text-sm text-muted-foreground mb-2">{subtitle}</div>
        )}
        {description && (
          <p className="text-xs text-muted-foreground">{description}</p>
        )}
        {trend && (
          <div className="flex items-center mt-3">
            <TrendingUp
              className={`h-4 w-4 mr-1 ${
                trend.isPositive ? "text-green-600" : "text-red-600"
              }`}
            />
            <span
              className={`text-xs font-medium ${
                trend.isPositive ? "text-green-600" : "text-red-600"
              }`}
            >
              {trend.isPositive ? "+" : ""}
              {trend.value}%
            </span>
          </div>
        )}
      </CardContent>
    </Card>
  );

  return to ? (
    <Link to={to} className="block">
      {content}
    </Link>
  ) : (
    content
  );
} 