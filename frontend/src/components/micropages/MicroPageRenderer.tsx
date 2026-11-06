'use client';

import React from 'react';
import { MicroPage, MicroLinkItem, MicroSocialLink, VCardContact, MicroPageTheme } from '@/types/micropage';

interface MicroPageRendererProps {
  page: Partial<MicroPage>;
  isPreview?: boolean;
  onLinkClick?: (linkId: string, url: string) => void;
  onSaveContact?: () => void;
  className?: string;
}

// Inline SVG social icons for zero dependency reliability
function SocialIcon({ platform }: { platform: string }) {
  const p = platform.toLowerCase();
  switch (p) {
    case 'twitter':
    case 'x':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M18.244 2.25h3.308l-7.227 8.26 8.502 11.24H16.17l-5.214-6.817L4.99 21.75H1.68l7.73-8.835L1.254 2.25H8.08l4.713 6.231zm-1.161 17.52h1.833L7.084 4.126H5.117z" />
        </svg>
      );
    case 'linkedin':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M19 3a2 2 0 0 1 2 2v14a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h14m-.5 15.5v-5.3a3.26 3.26 0 0 0-3.26-3.26c-.85 0-1.84.52-2.28 1.3v-1.11h-2.79v8.37h2.79v-4.93c0-.77.62-1.4 1.39-1.4a1.4 1.4 0 0 1 1.4 1.4v4.93h2.75M6.46 10.9h2.79v8.37H6.46v-8.37M7.86 6.88a1.45 1.45 0 0 0-1.46 1.46 1.46 1.46 0 0 0 1.46 1.46 1.46 1.46 0 0 0 1.46-1.46c0-.81-.66-1.46-1.46-1.46z" />
        </svg>
      );
    case 'github':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M12 2A10 10 0 0 0 2 12c0 4.42 2.87 8.17 6.84 9.5.5.08.66-.23.66-.5v-1.69c-2.77.6-3.36-1.34-3.36-1.34-.46-1.16-1.11-1.47-1.11-1.47-.91-.62.07-.6.07-.6 1 .07 1.53 1.03 1.53 1.03.87 1.52 2.34 1.07 2.91.83.1-.65.35-1.09.63-1.34-2.22-.25-4.55-1.11-4.55-4.92 0-1.11.38-2 1.03-2.71-.1-.25-.45-1.29.1-2.64 0 0 .84-.27 2.75 1.02.79-.22 1.65-.33 2.5-.33.85 0 1.71.11 2.5.33 1.91-1.29 2.75-1.02 2.75-1.02.55 1.35.2 2.39.1 2.64.65.71 1.03 1.6 1.03 2.71 0 3.82-2.34 4.66-4.57 4.91.36.31.69.92.69 1.85V21c0 .27.16.59.67.5C19.14 20.16 22 16.42 22 12A10 10 0 0 0 12 2z" />
        </svg>
      );
    case 'instagram':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M12 2.163c3.204 0 3.584.012 4.85.07 3.252.148 4.771 1.691 4.919 4.919.058 1.265.069 1.645.069 4.849 0 3.205-.012 3.584-.069 4.849-.149 3.225-1.664 4.771-4.919 4.919-1.266.058-1.644.07-4.85.07-3.204 0-3.584-.012-4.849-.07-3.26-.149-4.771-1.699-4.919-4.92-.058-1.265-.07-1.644-.07-4.849 0-3.204.013-3.583.07-4.849.149-3.227 1.664-4.771 4.919-4.919 1.266-.057 1.645-.069 4.849-.069zm0-2.163c-3.259 0-3.667.014-4.947.072-4.358.2-6.78 2.618-6.98 6.98-.059 1.281-.073 1.689-.073 4.948 0 3.259.014 3.668.072 4.948.2 4.358 2.618 6.78 6.98 6.98 1.281.058 1.689.072 4.948.072 3.259 0 3.668-.014 4.948-.072 4.354-.2 6.782-2.618 6.979-6.98.059-1.28.073-1.689.073-4.948 0-3.259-.014-3.667-.072-4.947-.196-4.354-2.617-6.78-6.979-6.98-1.281-.059-1.69-.073-4.949-.073zm0 5.838a6.162 6.162 0 1 0 0 12.324 6.162 6.162 0 0 0 0-12.324zm0 10.162a3.999 3.999 0 1 1 0-7.998 3.999 3.999 0 0 1 0 7.998zm6.406-11.845a1.44 1.44 0 1 0 0 2.881 1.44 1.44 0 0 0 0-2.881z" />
        </svg>
      );
    case 'youtube':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M23.498 6.186a3.016 3.016 0 0 0-2.122-2.136C19.505 3.545 12 3.545 12 3.545s-7.505 0-9.377.505A3.017 3.017 0 0 0 .502 6.186C0 8.07 0 12 0 12s0 3.93.502 5.814a3.016 3.016 0 0 0 2.122 2.136c1.871.505 9.376.505 9.376.505s7.505 0 9.377-.505a3.015 3.015 0 0 0 2.122-2.136C24 15.93 24 12 24 12s0-3.93-.502-5.814zM9.545 15.568V8.432L15.818 12l-6.273 3.568z" />
        </svg>
      );
    case 'whatsapp':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M17.472 14.382c-.301-.15-1.78-.877-2.056-.977-.276-.1-.477-.15-.678.15-.201.3-.778.977-.954 1.178-.175.2-.351.226-.652.075-.301-.15-1.272-.469-2.423-1.496-.895-.798-1.5-1.784-1.675-2.085-.176-.301-.019-.464.132-.614.136-.135.301-.351.451-.527.151-.175.201-.3.302-.501.1-.2.05-.376-.025-.526-.075-.15-.677-1.631-.928-2.233-.244-.586-.492-.506-.677-.515-.175-.009-.376-.011-.577-.011s-.527.075-.803.376c-.276.301-1.054 1.03-1.054 2.511 0 1.482 1.079 2.912 1.23 3.113.151.2 2.124 3.242 5.145 4.547.719.311 1.28.497 1.718.636.722.23 1.378.197 1.898.12.579-.087 1.78-.727 2.031-1.43.251-.703.251-1.305.176-1.43-.076-.126-.277-.201-.578-.352zM12 2C6.477 2 2 6.477 2 12c0 1.891.524 3.662 1.434 5.176L2.05 22l4.975-1.35A9.957 9.957 0 0 0 12 22c5.523 0 10-4.477 10-10S17.523 2 12 2z" />
        </svg>
      );
    case 'telegram':
      return (
        <svg className="w-5 h-5 fill-current" viewBox="0 0 24 24">
          <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm4.64 6.8c-.15 1.58-.8 5.42-1.13 7.19-.14.75-.42 1-.68 1.03-.58.05-1.02-.38-1.58-.75-.88-.58-1.38-.94-2.23-1.5-.99-.65-.35-1.01.22-1.59.15-.15 2.71-2.48 2.76-2.69.01-.03.01-.14-.06-.19-.06-.06-.17-.04-.24-.02-.11.02-1.78 1.13-5.02 3.32-.47.33-.9.49-1.29.48-.42-.01-1.24-.24-1.84-.44-.75-.24-1.34-.37-1.29-.79.03-.22.33-.44.9-.68 3.52-1.53 5.87-2.54 7.05-3.03 3.36-1.4 4.06-1.64 4.51-1.65.1 0 .32.02.46.14.12.1.15.24.17.34-.01.07.01.21 0 .28z" />
        </svg>
      );
    case 'email':
    case 'mail':
      return (
        <svg className="w-5 h-5 fill-none stroke-current" strokeWidth="2" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z" />
        </svg>
      );
    case 'phone':
      return (
        <svg className="w-5 h-5 fill-none stroke-current" strokeWidth="2" viewBox="0 0 24 24">
          <path strokeLinecap="round" strokeLinejoin="round" d="M3 5a2 2 0 012-2h3.28a1 1 0 01.948.684l1.498 4.493a1 1 0 01-.502 1.21l-2.257 1.13a11.042 11.042 0 005.516 5.516l1.13-2.257a1 1 0 011.21-.502l4.493 1.498a1 1 0 01.684.949V19a2 2 0 01-2 2h-1C9.716 21 3 14.284 3 6V5z" />
        </svg>
      );
    default:
      return (
        <svg className="w-5 h-5 fill-none stroke-current" strokeWidth="2" viewBox="0 0 24 24">
          <circle cx="12" cy="12" r="10" />
          <path strokeLinecap="round" strokeLinejoin="round" d="M12 2a15.3 15.3 0 0 1 4 10 15.3 15.3 0 0 1-4 10 15.3 15.3 0 0 1-4-10 15.3 15.3 0 0 1 4-10zM2 12h20" />
        </svg>
      );
  }
}

export function MicroPageRenderer({
  page,
  isPreview = false,
  onLinkClick,
  onSaveContact,
  className = '',
}: MicroPageRendererProps) {
  const theme: MicroPageTheme = page.theme || 'executive-obsidian';
  const vcard = page.vcard;
  const links: MicroLinkItem[] = page.links || [];
  const socialLinks: MicroSocialLink[] = page.social_links || [];

  const handleLinkClick = (e: React.MouseEvent, link: MicroLinkItem) => {
    if (isPreview) {
      e.preventDefault();
    }
    if (onLinkClick) {
      onLinkClick(link.id, link.url);
    }
  };

  const handleSaveContact = () => {
    if (onSaveContact) {
      onSaveContact();
    }
  };

  // Render template styles
  const getThemeWrapperClass = () => {
    switch (theme) {
      case 'executive-obsidian':
        return 'bg-[#0B0F17] text-slate-100 font-sans';
      case 'california-minimalist':
        return 'bg-[#FAF7F2] text-[#2D2A26] font-sans';
      case 'cyber-neon':
        return 'bg-[#070913] text-cyan-300 font-mono';
      case 'nordic-sage':
        return 'bg-[#F3F6F4] text-[#243329] font-sans';
      case 'swiss-grid':
        return 'bg-white text-black font-sans';
      case 'sunset-glass':
        return 'bg-gradient-to-br from-indigo-950 via-purple-900 to-rose-900 text-white font-sans';
      case 'cobalt-steel':
        return 'bg-[#0F172A] text-slate-100 font-sans';
      case 'editorial-monocle':
        return 'bg-[#FAF9F5] text-[#1A1A1A] font-serif';
      case 'neo-brutalist':
        return 'bg-[#FFFDF0] text-black font-sans';
      case 'aura-holographic':
        return 'bg-gradient-to-tr from-[#13111C] via-[#1F1934] to-[#121B2C] text-white font-sans';
      default:
        return 'bg-[#0B0F17] text-slate-100';
    }
  };

  const getSaveButtonClass = () => {
    switch (theme) {
      case 'executive-obsidian':
        return 'bg-gradient-to-r from-amber-500 to-amber-600 hover:from-amber-400 hover:to-amber-500 text-black font-semibold shadow-lg shadow-amber-500/20 rounded-xl';
      case 'california-minimalist':
        return 'bg-[#C66A48] hover:bg-[#B35938] text-white font-medium shadow-md shadow-[#C66A48]/20 rounded-2xl';
      case 'cyber-neon':
        return 'bg-gradient-to-r from-cyan-500 to-fuchsia-600 hover:from-cyan-400 hover:to-fuchsia-500 text-black font-bold uppercase tracking-wider shadow-[0_0_20px_rgba(6,182,212,0.5)] rounded-lg';
      case 'nordic-sage':
        return 'bg-[#3D5A47] hover:bg-[#2E4535] text-white font-medium shadow-md rounded-full';
      case 'swiss-grid':
        return 'bg-[#002FA7] hover:bg-black text-white font-bold tracking-tight uppercase rounded-none border-2 border-black';
      case 'sunset-glass':
        return 'bg-gradient-to-r from-rose-500 to-amber-500 hover:from-rose-400 hover:to-amber-400 text-white font-semibold shadow-lg shadow-rose-500/30 rounded-2xl';
      case 'cobalt-steel':
        return 'bg-blue-600 hover:bg-blue-500 text-white font-medium shadow-md shadow-blue-500/25 rounded-xl';
      case 'editorial-monocle':
        return 'bg-[#1A1A1A] hover:bg-black text-[#FAF9F5] uppercase tracking-widest text-xs font-serif rounded-none';
      case 'neo-brutalist':
        return 'bg-[#CCFF00] hover:bg-[#b8e600] text-black font-black uppercase tracking-wider border-[3px] border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] rounded-xl';
      case 'aura-holographic':
        return 'bg-gradient-to-r from-teal-400 via-purple-500 to-pink-500 hover:opacity-95 text-white font-semibold shadow-[0_0_25px_rgba(168,85,247,0.4)] rounded-2xl';
      default:
        return 'bg-amber-500 text-black font-medium rounded-xl';
    }
  };

  const getCardClass = (isHighlighted = false) => {
    switch (theme) {
      case 'executive-obsidian':
        return isHighlighted
          ? 'bg-[#1C2128] border-2 border-amber-500/60 shadow-amber-500/10 shadow-lg text-white rounded-xl'
          : 'bg-[#161B22]/80 backdrop-blur-md border border-[#30363D] hover:border-amber-500/50 hover:bg-[#1C2128] text-slate-200 rounded-xl shadow-md';
      case 'california-minimalist':
        return isHighlighted
          ? 'bg-white border-2 border-[#C66A48]/50 shadow-md text-[#2D2A26] rounded-2xl'
          : 'bg-white/80 border border-[#E8E2D9] hover:border-[#D4CBBF] hover:bg-white text-[#2D2A26] rounded-2xl shadow-sm hover:shadow';
      case 'cyber-neon':
        return isHighlighted
          ? 'bg-[#0e1224] border-2 border-fuchsia-500 text-fuchsia-300 shadow-[0_0_20px_rgba(217,70,239,0.3)] rounded-lg'
          : 'bg-[#0e1224]/80 border border-cyan-500/30 hover:border-cyan-400 text-cyan-200 hover:shadow-[0_0_15px_rgba(6,182,212,0.25)] rounded-lg';
      case 'nordic-sage':
        return isHighlighted
          ? 'bg-white border-2 border-[#3D5A47] shadow-md text-[#243329] rounded-3xl'
          : 'bg-white/90 border border-[#D8E2DC] hover:border-[#8FA396] text-[#243329] rounded-3xl shadow-sm hover:shadow';
      case 'swiss-grid':
        return isHighlighted
          ? 'bg-black text-white border-2 border-black rounded-none font-bold'
          : 'bg-white border-2 border-black hover:bg-slate-100 text-black rounded-none';
      case 'sunset-glass':
        return isHighlighted
          ? 'bg-white/20 backdrop-blur-xl border-2 border-rose-300/60 text-white rounded-2xl shadow-xl'
          : 'bg-white/10 backdrop-blur-xl border border-white/20 hover:bg-white/15 text-white rounded-2xl shadow-lg';
      case 'cobalt-steel':
        return isHighlighted
          ? 'bg-slate-800 border-2 border-blue-500 text-white shadow-lg shadow-blue-500/20 rounded-xl'
          : 'bg-slate-800/80 border border-slate-700 hover:border-blue-500/60 hover:bg-slate-800 text-slate-200 rounded-xl shadow';
      case 'editorial-monocle':
        return isHighlighted
          ? 'bg-[#1A1A1A]/5 border-y-2 border-[#1A1A1A] text-[#1A1A1A] rounded-none'
          : 'bg-transparent border-b border-[#1A1A1A]/20 hover:border-[#1A1A1A] text-[#1A1A1A] rounded-none hover:bg-black/5';
      case 'neo-brutalist':
        return isHighlighted
          ? 'bg-[#FFE600] border-[3px] border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] text-black rounded-xl font-bold'
          : 'bg-white border-[3px] border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] hover:translate-x-[2px] hover:translate-y-[2px] hover:shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] text-black rounded-xl font-bold';
      case 'aura-holographic':
        return isHighlighted
          ? 'bg-white/10 backdrop-blur-lg border-2 border-teal-300/60 text-white shadow-[0_0_20px_rgba(45,212,191,0.3)] rounded-2xl'
          : 'bg-white/5 backdrop-blur-lg border border-purple-400/20 hover:border-teal-300/40 hover:bg-white/10 text-white rounded-2xl shadow-lg';
      default:
        return 'bg-slate-800 border border-slate-700 rounded-xl';
    }
  };

  const getSocialIconContainerClass = () => {
    switch (theme) {
      case 'executive-obsidian':
        return 'bg-[#161B22] border border-[#30363D] hover:border-amber-500 text-slate-300 hover:text-amber-400 rounded-full';
      case 'california-minimalist':
        return 'bg-white border border-[#E8E2D9] hover:border-[#C66A48] text-[#2D2A26] hover:text-[#C66A48] rounded-full';
      case 'cyber-neon':
        return 'bg-[#0e1224] border border-cyan-500/40 hover:border-fuchsia-500 text-cyan-300 hover:text-fuchsia-400 rounded-lg shadow-[0_0_10px_rgba(6,182,212,0.2)]';
      case 'nordic-sage':
        return 'bg-white border border-[#D8E2DC] hover:border-[#3D5A47] text-[#243329] hover:text-[#3D5A47] rounded-full';
      case 'swiss-grid':
        return 'bg-white border-2 border-black hover:bg-black text-black hover:text-white rounded-none';
      case 'sunset-glass':
        return 'bg-white/10 backdrop-blur-md border border-white/20 hover:border-white/50 text-white hover:text-amber-200 rounded-2xl';
      case 'cobalt-steel':
        return 'bg-slate-800 border border-slate-700 hover:border-blue-500 text-slate-300 hover:text-blue-400 rounded-xl';
      case 'editorial-monocle':
        return 'bg-transparent border border-[#1A1A1A]/30 hover:border-[#1A1A1A] text-[#1A1A1A] rounded-none';
      case 'neo-brutalist':
        return 'bg-white border-[2px] border-black shadow-[2px_2px_0px_0px_rgba(0,0,0,1)] text-black hover:bg-[#CCFF00] rounded-lg';
      case 'aura-holographic':
        return 'bg-white/5 backdrop-blur-md border border-purple-400/30 hover:border-teal-300 text-purple-200 hover:text-teal-200 rounded-xl';
      default:
        return 'bg-slate-800 border border-slate-700 text-slate-200 rounded-full';
    }
  };

  const getAvatarBorderClass = () => {
    switch (theme) {
      case 'executive-obsidian':
        return 'border-2 border-amber-500/60 shadow-lg shadow-amber-500/20';
      case 'california-minimalist':
        return 'border-2 border-white shadow-md';
      case 'cyber-neon':
        return 'border-2 border-cyan-400 shadow-[0_0_20px_rgba(6,182,212,0.6)] rounded-xl';
      case 'nordic-sage':
        return 'border-2 border-white shadow-sm';
      case 'swiss-grid':
        return 'border-3 border-black rounded-none';
      case 'sunset-glass':
        return 'border-2 border-white/40 shadow-xl';
      case 'cobalt-steel':
        return 'border-2 border-blue-500/80 shadow-md';
      case 'editorial-monocle':
        return 'border border-[#1A1A1A] rounded-none p-1 bg-white';
      case 'neo-brutalist':
        return 'border-[3px] border-black shadow-[4px_4px_0px_0px_rgba(0,0,0,1)] rounded-xl';
      case 'aura-holographic':
        return 'border-2 border-teal-300/80 shadow-[0_0_25px_rgba(168,85,247,0.5)]';
      default:
        return 'border-2 border-slate-600';
    }
  };

  return (
    <div className={`w-full min-h-full flex flex-col justify-between ${getThemeWrapperClass()} ${className}`}>
      {/* Cover Banner */}
      <div className="relative w-full h-36 md:h-44 overflow-hidden bg-slate-800/40 flex-shrink-0">
        {page.cover_url ? (
          <img
            src={page.cover_url}
            alt="Cover"
            className="w-full h-full object-cover"
          />
        ) : (
          <div className="w-full h-full opacity-40 bg-[radial-gradient(ellipse_at_top,_var(--tw-gradient-stops))] from-white/20 via-transparent to-transparent" />
        )}
      </div>

      {/* Profile & vCard Content */}
      <div className="relative px-6 -mt-16 flex-1 flex flex-col items-center max-w-lg mx-auto w-full pb-10">
        {/* Avatar */}
        <div className="relative mb-4">
          {page.avatar_url ? (
            <img
              src={page.avatar_url}
              alt={page.title || 'Profile'}
              className={`w-28 h-28 object-cover rounded-full ${getAvatarBorderClass()}`}
            />
          ) : (
            <div
              className={`w-28 h-28 rounded-full flex items-center justify-center text-3xl font-bold uppercase ${getAvatarBorderClass()} bg-slate-700 text-white`}
            >
              {(page.title || 'V')[0]}
            </div>
          )}
        </div>

        {/* Identity Headings */}
        <h1 className="text-2xl font-bold text-center tracking-tight leading-snug">
          {page.title || 'Your Name'}
        </h1>

        {/* Job Title & Organization */}
        {(vcard?.title || vcard?.organization) && (
          <div className="mt-1 text-sm text-center opacity-80 flex items-center gap-1.5 flex-wrap justify-center">
            {vcard.title && <span className="font-medium">{vcard.title}</span>}
            {vcard.title && vcard.organization && <span>•</span>}
            {vcard.organization && <span>{vcard.organization}</span>}
          </div>
        )}

        {/* Bio */}
        {page.bio && (
          <p className="mt-3 text-sm text-center opacity-75 max-w-sm leading-relaxed">
            {page.bio}
          </p>
        )}

        {/* Primary Action: Save Contact (.vcf) */}
        <div className="w-full mt-5">
          <button
            type="button"
            onClick={handleSaveContact}
            className={`w-full py-3.5 px-6 flex items-center justify-center gap-2.5 transition-all active:scale-[0.98] ${getSaveButtonClass()}`}
          >
            <svg className="w-5 h-5 fill-none stroke-current" strokeWidth="2.2" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-4l-4 4m0 0l-4-4m4 4V4" />
            </svg>
            <span>Save Contact to Phone</span>
          </button>
        </div>

        {/* Quick Contact Details Pills */}
        {vcard && (vcard.email || vcard.phone || vcard.city || vcard.website) && (
          <div className="mt-4 flex flex-wrap gap-2 justify-center w-full">
            {vcard.email && (
              <a
                href={isPreview ? '#' : `mailto:${vcard.email}`}
                className="text-xs px-3 py-1.5 rounded-full border border-current/20 hover:border-current/40 flex items-center gap-1.5 transition-opacity opacity-85 hover:opacity-100"
              >
                <SocialIcon platform="email" />
                <span className="truncate max-w-[180px]">{vcard.email}</span>
              </a>
            )}
            {vcard.phone && (
              <a
                href={isPreview ? '#' : `tel:${vcard.phone}`}
                className="text-xs px-3 py-1.5 rounded-full border border-current/20 hover:border-current/40 flex items-center gap-1.5 transition-opacity opacity-85 hover:opacity-100"
              >
                <SocialIcon platform="phone" />
                <span>{vcard.phone}</span>
              </a>
            )}
            {vcard.city && (
              <span className="text-xs px-3 py-1.5 rounded-full border border-current/20 flex items-center gap-1.5 opacity-85">
                <svg className="w-4 h-4 fill-none stroke-current" strokeWidth="2" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" d="M17.657 16.657L13.414 20.9a1.998 1.998 0 01-2.827 0l-4.244-4.243a8 8 0 1111.314 0z" />
                  <path strokeLinecap="round" strokeLinejoin="round" d="M15 11a3 3 0 11-6 0 3 3 0 016 0z" />
                </svg>
                <span>{vcard.city}{vcard.state ? `, ${vcard.state}` : ''}</span>
              </span>
            )}
          </div>
        )}

        {/* Social Profile Icons */}
        {socialLinks.length > 0 && (
          <div className="mt-5 flex items-center justify-center gap-2.5 flex-wrap">
            {socialLinks.map((s, idx) => (
              <a
                key={idx}
                href={isPreview ? '#' : s.url}
                target={isPreview ? '_self' : '_blank'}
                rel="noopener noreferrer"
                title={s.label || s.platform}
                className={`w-10 h-10 flex items-center justify-center transition-all hover:scale-105 active:scale-95 ${getSocialIconContainerClass()}`}
              >
                <SocialIcon platform={s.platform} />
              </a>
            ))}
          </div>
        )}

        {/* Interactive Custom Links */}
        <div className="w-full mt-6 space-y-3">
          {links.filter((l) => l.is_active !== false).map((link) => (
            <a
              key={link.id}
              href={isPreview ? '#' : link.url}
              target={isPreview ? '_self' : '_blank'}
              rel="noopener noreferrer"
              onClick={(e) => handleLinkClick(e, link)}
              className={`block w-full p-4 transition-all hover:scale-[1.01] active:scale-[0.99] cursor-pointer ${getCardClass(link.is_highlighted)}`}
            >
              <div className="flex items-center justify-between">
                <div className="flex items-center gap-3 min-w-0">
                  {link.icon && (
                    <span className="text-xl flex-shrink-0">{link.icon}</span>
                  )}
                  <div className="min-w-0">
                    <div className="font-semibold text-sm truncate">{link.title}</div>
                    {link.description && (
                      <div className="text-xs opacity-70 truncate mt-0.5">{link.description}</div>
                    )}
                  </div>
                </div>
                <svg className="w-4 h-4 opacity-50 flex-shrink-0 ml-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth="2.5" d="M9 5l7 7-7 7" />
                </svg>
              </div>
            </a>
          ))}
        </div>
      </div>

      {/* Footer Branding */}
      <footer className="py-6 text-center text-xs opacity-60">
        <a
          href="/"
          target="_blank"
          rel="noopener noreferrer"
          className="hover:underline inline-flex items-center gap-1.5"
        >
          <span>Created with</span>
          <span className="font-bold tracking-tight">Trunc vCard</span>
        </a>
      </footer>
    </div>
  );
}
