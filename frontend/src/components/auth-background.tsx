"use client"

import { useState, useEffect, useCallback } from "react"

const phrases = [
  "Focus on findings that matter",
  "Track vulnerabilities across your stack",
  "Monitor your application security posture",
  "Secure your deployment pipeline",
  "Stay compliant with automated policies",
]

type Phase = "typing" | "waiting" | "fading"

export function AuthBackground() {
  const [index, setIndex] = useState(0)
  const [chars, setChars] = useState(0)
  const [phase, setPhase] = useState<Phase>("typing")
  const [opacity, setOpacity] = useState(1)

  const advance = useCallback(() => {
    setIndex((prev) => (prev + 1) % phrases.length)
    setChars(0)
    setOpacity(1)
    setPhase("typing")
  }, [])

  useEffect(() => {
    if (phase === "typing" && chars < phrases[index].length) {
      const t = setTimeout(() => setChars((c) => c + 1), 40)
      return () => clearTimeout(t)
    }
    if (phase === "typing" && chars === phrases[index].length) {
      setPhase("waiting")
    }
  }, [phase, chars, index])

  useEffect(() => {
    if (phase !== "waiting") return
    const t = setTimeout(() => setPhase("fading"), 2800)
    return () => clearTimeout(t)
  }, [phase])

  useEffect(() => {
    if (phase !== "fading") return
    setOpacity(0)
    const t = setTimeout(advance, 500)
    return () => clearTimeout(t)
  }, [phase, advance])

  return (
    <div className="relative h-full w-full overflow-hidden bg-gradient-to-br from-primary/[0.15] via-card to-accent/[0.10]">
      <div className="absolute -top-1/2 -left-1/3 w-3/4 h-3/4 rounded-full bg-[#A298AB]/15 blur-[120px]" />
      <div className="absolute -bottom-1/2 -right-1/3 w-3/4 h-3/4 rounded-full bg-[#8F849A]/12 blur-[120px]" />
      <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-[30%] h-[40%] rounded-full bg-[#B4ADBC]/6 blur-[100px]" />

      <div className="absolute inset-0 z-10 flex flex-col items-center justify-center p-16">
        <div className="flex flex-col items-center gap-8 max-w-md text-center">
          <img
            src="/assets/servasec-mark.svg"
            alt="servasec"
            className="w-36 h-36 opacity-80"
          />
          <p
            className="text-lg leading-relaxed text-foreground/75 font-light tracking-wide transition-opacity duration-500 ease-in-out min-h-[1.75em]"
            style={{ opacity }}
          >
            {phrases[index].slice(0, chars)}
            {chars < phrases[index].length && (
              <span className="inline-block w-[2px] h-[1.1em] bg-foreground/60 ml-0.5 align-middle animate-pulse" />
            )}
          </p>
        </div>
      </div>
    </div>
  )
}
