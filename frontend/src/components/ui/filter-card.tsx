import type { ReactNode } from "react";
import type { LucideIcon } from "lucide-react";
import { Card, CardContent, CardHeader, CardTitle } from "./card";
import { cn } from "@/lib/utils";

interface FilterCardProps {
  title?: string;
  icon?: LucideIcon;
  children: ReactNode;
  className?: string;
  contentClassName?: string;
}

export function FilterCard({
  title = "筛选和搜索",
  icon: Icon,
  children,
  className,
  contentClassName,
}: FilterCardProps) {
  return (
    <Card className={cn("rounded-xl shadow-lg", className)}>
      <CardHeader>
        <CardTitle className="flex items-center gap-2">
          {Icon && <Icon className="h-5 w-5" />}
          {title}
        </CardTitle>
      </CardHeader>
      <CardContent className={cn("space-y-4", contentClassName)}>
        {children}
      </CardContent>
    </Card>
  );
}

