"use client"

import { useTheme } from "@/contexts/ThemeContext"
import { useEffect, useState } from "react"

export function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const [isDark, setIsDark] = useState(false)

  useEffect(() => {
    if (theme === "system") {
      const systemTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
      setIsDark(systemTheme)
    } else {
      setIsDark(theme === "dark")
    }
  }, [theme])

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