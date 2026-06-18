"use client"

export function AuthBackground() {
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
          <p className="text-lg leading-relaxed text-foreground/75 font-light tracking-wide">
            Focus on findings that matter
          </p>
        </div>
      </div>
    </div>
  )
}
