import { StatCard } from "./stat-card";
import type { LucideIcon } from "lucide-react";

interface StatItem {
  title: string;
  value: string | number;
  icon: LucideIcon;
  color?: "info" | "success" | "warning" | "danger" | "purple";
  subtitle?: string;
}

interface StatsGridProps {
  stats: StatItem[];
  className?: string;
  columns?: 2 | 3 | 4 | 5 | 6;
}

export function StatsGrid({ 
  stats, 
  className = "", 
  columns = 4 
}: StatsGridProps) {
  const gridCols = {
    2: "md:grid-cols-2",
    3: "md:grid-cols-3", 
    4: "md:grid-cols-4",
    5: "md:grid-cols-5",
    6: "md:grid-cols-6"
  };

  return (
    <div className={`grid gap-4 ${gridCols[columns]} ${className}`}>
      {stats.map((stat, index) => (
        <StatCard
          key={index}
          title={stat.title}
          value={stat.value}
          icon={stat.icon}
          color={stat.color}
          subtitle={stat.subtitle}
        />
      ))}
    </div>
  );
} 