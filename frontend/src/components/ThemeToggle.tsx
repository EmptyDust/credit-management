"use client"

import { useTheme } from "@/contexts/ThemeContext"

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const isDark =
    theme === "system"
      ? window.matchMedia("(prefers-color-scheme: dark)").matches
      : theme === "dark"

  const handleToggle = () => {
    const newTheme = isDark ? "light" : "dark"
    setTheme(newTheme)
  }

  return (
    <div className="toggle-switch">
      <label className="switch-label">
        <input
          type="checkbox"
          className="checkbox"
          checked={isDark}
          onChange={handleToggle}
        />
        <span className="slider"></span>
      </label>
    </div>
  )
}
