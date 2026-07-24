import { useState, useEffect, useRef } from "react";
import { motion, AnimatePresence } from "motion/react";
import { useTheme } from "next-themes";
import { Dialog, DialogContent } from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Sparkles, ArrowRight, ArrowLeft, X } from "lucide-react";
import api from "@/lib/api";
import type { User } from "@/lib/types";

interface Slide {
  title: string;
  description: string;
  image?: { light: string; dark: string };
  isLogo?: boolean;
}

const slides: Slide[] = [
  {
    title: "Welcome to ServaSec",
    description:
      "Your vulnerability management platform. Centralize, analyze, and track security findings across all your applications.",
    isLogo: true,
  },
  {
    title: "Organize with Groups",
    description:
      "Create groups to structure your applications by project, team, or environment. Keep everything tidy and accessible.",
    image: {
      light: "/assets/screenshots/02-dashboard.png",
      dark: "/assets/screenshots/02-dashboard-dark.png",
    },
  },
  {
    title: "Manage Applications",
    description:
      "Add your applications, trigger scans, and track version history. Every scan is stored and comparable over time.",
    image: {
      light: "/assets/screenshots/04-applications.png",
      dark: "/assets/screenshots/04-applications-dark.png",
    },
  },
  {
    title: "Track Findings",
    description:
      "Browse, filter, and sort vulnerabilities. Follow each finding through its lifecycle from detection to resolution.",
    image: {
      light: "/assets/screenshots/03-findings.png",
      dark: "/assets/screenshots/03-findings-dark.png",
    },
  },
  {
    title: "Ready to go?",
    description:
      "Create your first group, add an application, and run a scan. You are all set!",
  },
];

interface OnboardingModalProps {
  open: boolean;
  onComplete: (user: User) => void;
  user: User;
}

export function OnboardingModal({ open, onComplete, user }: OnboardingModalProps) {
  const { resolvedTheme } = useTheme();
  const [step, setStep] = useState(0);
  const [direction, setDirection] = useState(0);
  const preloaded = useRef(new Set<string>());

  const slide = slides[step];
  const isLast = step === slides.length - 1;
  const isFirst = step === 0;

  const isDark = resolvedTheme === "dark" || resolvedTheme === "catppuccin" || resolvedTheme === "github" || resolvedTheme === "nord";
  const markSrc = "/assets/servasec-mark.svg";
  const imgSrc = slide.image
    ? isDark && slide.image.dark
      ? slide.image.dark
      : slide.image.light
    : null;

  // Preload adjacent images
  useEffect(() => {
    slides.forEach((s, i) => {
      if (!s.image) return;
      const src = isDark ? s.image.dark : s.image.light;
      if (src && !preloaded.current.has(src)) {
        const img = new Image();
        img.src = src;
        preloaded.current.add(src);
      }
    });
  }, [isDark]);

  const markSeen = async () => {
    if (user.hasSeenOnboarding) return;
    try {
      await api.patch("/api/me/onboarding");
      onComplete({ ...user, hasSeenOnboarding: true });
    } catch {
      onComplete(user);
    }
  };

  const handleClose = () => {
    markSeen();
  };

  const handleNext = () => {
    if (isLast) {
      markSeen();
    } else {
      setDirection(1);
      setStep((s) => s + 1);
    }
  };

  const handlePrev = () => {
    if (isFirst) return;
    setDirection(-1);
    setStep((s) => s - 1);
  };

  const slideVariants = {
    enter: (d: number) => ({ x: d > 0 ? 60 : -60, opacity: 0 }),
    center: { x: 0, opacity: 1 },
    exit: (d: number) => ({ x: d > 0 ? -60 : 60, opacity: 0 }),
  };

  return (
    <Dialog open={open} onOpenChange={(v) => { if (!v) handleClose(); }}>
      <DialogContent
        className="sm:max-w-md p-0 gap-0 overflow-hidden [&>button:last-child]:hidden"
        onInteractOutside={(e) => e.preventDefault()}
        onPointerDownOutside={(e) => e.preventDefault()}
      >
        <div className="relative px-6 pt-8 pb-6">
          <button
            onClick={handleClose}
            className="absolute right-4 top-4 rounded-sm opacity-60 hover:opacity-100 transition-opacity text-muted-foreground z-10"
            aria-label="Skip"
          >
            <X className="h-4 w-4" />
          </button>

          {/* Progress dots */}
          <div className="flex items-center justify-center gap-1.5 mb-5">
            {slides.map((_, i) => (
              <div
                key={i}
                className="h-1 rounded-full transition-all duration-300 ease-out"
                style={{
                  width: i === step ? 20 : 5,
                  backgroundColor: i === step
                    ? "hsl(var(--primary))"
                    : "hsl(var(--muted-foreground) / 0.2)",
                }}
              />
            ))}
          </div>

          {/* Slide content */}
          <div className="relative min-h-[16rem]">
            <AnimatePresence mode="popLayout" custom={direction}>
              <motion.div
                key={step}
                custom={direction}
                variants={slideVariants}
                initial="enter"
                animate="center"
                exit="exit"
                transition={{ duration: 0.25, ease: [0.25, 0.1, 0.25, 1] }}
                className="w-full"
              >
                {slide.isLogo ? (
                  <div className="flex flex-col items-center text-center pt-4">
                    <div className="mb-5">
                      <div className="w-16 h-16 rounded-2xl bg-primary/10 flex items-center justify-center">
                        <img src={markSrc} alt="ServaSec" className="w-9 h-9" />
                      </div>
                    </div>
                    <h2 className="text-lg font-semibold mb-2">{slide.title}</h2>
                    <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                      {slide.description}
                    </p>
                  </div>
                ) : imgSrc ? (
                  <div className="flex flex-col items-center text-center">
                    <img
                      src={imgSrc}
                      alt={slide.title}
                      className="w-full rounded-lg border border-border mb-4"
                    />
                    <h2 className="text-lg font-semibold mb-1.5">{slide.title}</h2>
                    <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                      {slide.description}
                    </p>
                  </div>
                ) : (
                  <div className="flex flex-col items-center text-center pt-8">
                    <div className="mb-4">
                      <div className="w-14 h-14 rounded-2xl bg-primary/10 flex items-center justify-center">
                        <Sparkles className="h-7 w-7 text-primary" />
                      </div>
                    </div>
                    <h2 className="text-lg font-semibold mb-2">{slide.title}</h2>
                    <p className="text-sm text-muted-foreground max-w-xs leading-relaxed">
                      {slide.description}
                    </p>
                  </div>
                )}
              </motion.div>
            </AnimatePresence>
          </div>

          {/* Navigation */}
          <div className="flex items-center justify-between mt-6">
            <div>
              {!isFirst && (
                <Button variant="ghost" size="sm" onClick={handlePrev}>
                  <ArrowLeft className="h-4 w-4 mr-1" />
                  Back
                </Button>
              )}
            </div>
            <div className="flex items-center gap-2">
              <Button variant="ghost" size="sm" onClick={handleClose}>
                Skip
              </Button>
              <Button size="sm" onClick={handleNext}>
                {isLast ? "Get started" : "Next"}
                {!isLast && <ArrowRight className="h-4 w-4 ml-1" />}
              </Button>
            </div>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
