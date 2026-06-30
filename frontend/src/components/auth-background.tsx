"use client"

import { useState, useEffect } from "react"
import { motion, AnimatePresence, useReducedMotion } from "motion/react"

const phrases = [
  "Focus on findings that matter",
  "Track vulnerabilities across your stack",
  "Monitor your application security posture",
  "Secure your deployment pipeline",
  "Stay compliant with automated policies",
]

const letters = "servasec".split("")

const containerVariants = (reduce: boolean) => ({
  hidden: {},
  visible: {
    transition: {
      staggerChildren: reduce ? 0 : 0.1,
      delayChildren: reduce ? 0 : 0.2,
    },
  },
  exit: {
    transition: {
      staggerChildren: reduce ? 0 : 0.03,
      staggerDirection: -1,
    },
  },
})

const letterVariants = (reduce: boolean) => ({
  hidden: { opacity: 0, x: reduce ? 0 : -30 },
  visible: {
    opacity: 1,
    x: 0,
    transition: {
      duration: reduce ? 0 : 0.4,
      ease: [0.16, 1, 0.3, 1] as [number, number, number, number],
    },
  },
  exit: {
    opacity: 0,
    y: reduce ? 0 : -30,
    transition: { duration: reduce ? 0 : 0.3 },
  },
})

export function AuthBackground() {
  const prefersReduce = useReducedMotion()
  const reduce = prefersReduce ?? false

  const [showLogo, setShowLogo] = useState(true)
  const [phraseIndex, setPhraseIndex] = useState(0)
  const [key, setKey] = useState(0)
  const [chars, setChars] = useState(0)

  useEffect(() => {
    const duration = showLogo ? 3500 : 5000
    const t = setTimeout(() => {
      if (showLogo) {
        setShowLogo(false)
      } else {
        setShowLogo(true)
        setPhraseIndex((i) => (i + 1) % phrases.length)
        setKey((k) => k + 1)
      }
    }, duration)
    return () => clearTimeout(t)
  }, [showLogo])

  useEffect(() => {
    if (showLogo) return
    setChars(0)
  }, [showLogo])

  useEffect(() => {
    if (showLogo) return
    if (chars >= phrases[phraseIndex].length) return
    const t = setTimeout(() => setChars((c) => c + 1), 40)
    return () => clearTimeout(t)
  }, [showLogo, chars, phraseIndex])

  return (
    <div className="relative h-full w-full overflow-hidden bg-background">
      <svg className="absolute w-0 h-0" aria-hidden>
        <filter id="grain">
          <feTurbulence
            type="fractalNoise"
            baseFrequency="0.8"
            numOctaves="3"
            stitchTiles="stitch"
          />
          <feColorMatrix type="saturate" values="0" />
        </filter>
      </svg>

      <div className="absolute -top-1/2 -left-1/3 w-3/4 h-3/4 rounded-full bg-primary/[0.15] blur-[160px]" />
      <div className="absolute -bottom-1/2 -right-1/3 w-3/4 h-3/4 rounded-full bg-accent/[0.10] blur-[160px]" />

      <div className="absolute inset-0 z-10 flex items-center justify-center p-16">
        <div className="flex flex-col items-center gap-6 max-w-md text-center">
          <AnimatePresence mode="wait">
            {showLogo ? (
              <motion.h1
                key={`logo-${key}`}
                variants={containerVariants(reduce)}
                initial="hidden"
                animate="visible"
                exit="exit"
                className="text-6xl md:text-7xl lg:text-8xl font-bold tracking-tighter"
              >
                {letters.map((char, i) => (
                  <motion.span
                    key={i}
                    variants={letterVariants(reduce)}
                    className={`inline-block ${i < 5 ? "text-foreground" : "text-primary/60"}`}
                  >
                    {char}
                  </motion.span>
                ))}
              </motion.h1>
            ) : (
              <motion.p
                key={`phrase-${phraseIndex}`}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -12 }}
                transition={{
                  duration: reduce ? 0 : 0.5,
                  ease: [0.16, 1, 0.3, 1] as [number, number, number, number],
                }}
                className="text-xl leading-relaxed text-foreground/70 font-light tracking-wide min-h-[2em]"
              >
                {phrases[phraseIndex].slice(0, chars)}
                {!reduce && chars < phrases[phraseIndex].length && (
                  <span className="inline-block w-[2px] h-[1.1em] bg-foreground/50 ml-0.5 align-middle animate-pulse" />
                )}
              </motion.p>
            )}
          </AnimatePresence>
        </div>
      </div>

      <div className="absolute inset-0 z-20 pointer-events-none overflow-hidden mix-blend-overlay">
        <svg className="w-full h-full" aria-hidden>
          <rect width="100%" height="100%" filter="url(#grain)" opacity="0.08" />
        </svg>
      </div>
    </div>
  )
}
