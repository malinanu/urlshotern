#!/usr/bin/env python3
"""
Generate hero images for URL shortener blog posts using HuggingFace Inference API.
Saves PNG images at ~1200x630 to frontend/public/blog/

Usage:
    HF_TOKEN=hf_xxx python3 generate_blog_images.py

Requires a free HuggingFace account token from https://huggingface.co/settings/tokens
(Read access is enough; no special permissions needed for public models)
"""

import os
import sys
import time
import io

# ---------------------------------------------------------------------------
# Config
# ---------------------------------------------------------------------------
OUTPUT_DIR = "/home/malinjayaweera/dev/personal/urlshotern/frontend/public/blog"

IMAGES = [
    {
        "filename": "how-to-shorten-a-url.png",
        "prompt": (
            "Photorealistic split scene: left side shows a tangled overwhelming wall of text URL "
            "on a dark terminal screen, right side shows a clean glowing short link on a bright "
            "mobile phone screen. Sharp diagonal cut between sides with vivid blue electric light. "
            "Shallow depth of field, editorial tech photography style, wide 16:9 composition, "
            "no text overlays, cinematic lighting"
        ),
    },
    {
        "filename": "trunc-vs-bitly-tinyurl.png",
        "prompt": (
            "Three floating smartphone screens arranged in aerial flat-lay view on a matte dark "
            "surface. Each screen displays a different mobile app dashboard UI. The centre screen "
            "is larger, brighter, elevated with a vivid blue glow. The two side screens are "
            "slightly smaller and dimmer. Clean product photography, soft studio lighting, "
            "dark background, wide 16:9 composition, no text"
        ),
    },
    {
        "filename": "best-url-shorteners-2026.png",
        "prompt": (
            "Dramatic hero shot: a glowing trophy made of blue light stands on a dark podium "
            "surrounded by seven glowing floating chain link icons arranged in a semicircle like "
            "a leaderboard. Deep cinematic dark background, editorial style, wide angle 16:9, "
            "volumetric blue lighting, no text, award ceremony atmosphere"
        ),
    },
]

# Models to try in order (primary → fallbacks)
MODELS = [
    "black-forest-labs/FLUX.1-schnell",
    "stabilityai/stable-diffusion-xl-base-1.0",
    "runwayml/stable-diffusion-v1-5",
]

# Target dimensions (closest 16:9 at 64px granularity supported by FLUX)
WIDTH = 1216
HEIGHT = 640


# ---------------------------------------------------------------------------
# Auth
# ---------------------------------------------------------------------------
def get_hf_token() -> str:
    token = os.environ.get("HF_TOKEN", "").strip()
    if token:
        return token
    try:
        from huggingface_hub import get_token as hf_get_token
        t = hf_get_token()
        if t:
            return t
    except Exception:
        pass
    for path in [
        os.path.expanduser("~/.cache/huggingface/token"),
        os.path.expanduser("~/.huggingface/token"),
    ]:
        if os.path.exists(path):
            with open(path) as f:
                t = f.read().strip()
            if t:
                return t
    return ""


HF_TOKEN = get_hf_token()
if HF_TOKEN:
    print(f"Using HF token: {HF_TOKEN[:8]}...")
else:
    print("ERROR: No HF_TOKEN found.")
    print("  1. Get a free token at: https://huggingface.co/settings/tokens")
    print("  2. Then run:  HF_TOKEN=hf_xxx python3 generate_blog_images.py")
    sys.exit(1)


# ---------------------------------------------------------------------------
# Generation via huggingface_hub InferenceClient
# ---------------------------------------------------------------------------
def generate_with_client(model: str, prompt: str) -> bytes:
    """Use huggingface_hub InferenceClient (handles auth, retries, loading)."""
    from huggingface_hub import InferenceClient

    client = InferenceClient(token=HF_TOKEN)

    print(f"  InferenceClient -> {model}")

    if "FLUX" in model or "flux" in model:
        result = client.text_to_image(
            prompt,
            model=model,
            width=WIDTH,
            height=HEIGHT,
            num_inference_steps=4,
            guidance_scale=0.0,
        )
    else:
        result = client.text_to_image(
            prompt,
            model=model,
            width=WIDTH,
            height=HEIGHT,
            num_inference_steps=30,
            guidance_scale=7.5,
        )

    # result is a PIL Image
    buf = io.BytesIO()
    result.save(buf, format="PNG")
    return buf.getvalue()


def generate_image(image_cfg: dict) -> bool:
    filename = image_cfg["filename"]
    prompt = image_cfg["prompt"]
    out_path = os.path.join(OUTPUT_DIR, filename)

    print(f"\n{'='*64}")
    print(f"Generating: {filename}")
    print(f"Prompt: {prompt[:100]}...")

    for model in MODELS:
        print(f"\n  Trying: {model}")
        try:
            image_bytes = generate_with_client(model, prompt)
            with open(out_path, "wb") as f:
                f.write(image_bytes)
            size_kb = len(image_bytes) / 1024
            print(f"  Saved: {out_path}  ({size_kb:.1f} KB)")
            return True
        except Exception as e:
            print(f"  FAILED with {model}: {e}")
            continue

    print(f"  All models failed for {filename}")
    return False


# ---------------------------------------------------------------------------
# Main
# ---------------------------------------------------------------------------
def main():
    # Ensure Pillow is available (needed by InferenceClient.text_to_image)
    try:
        from PIL import Image  # noqa: F401
    except ImportError:
        print("Installing Pillow...")
        import subprocess
        subprocess.run(
            [sys.executable, "-m", "pip", "install", "--break-system-packages", "Pillow"],
            check=True, capture_output=True
        )

    os.makedirs(OUTPUT_DIR, exist_ok=True)

    results = []
    for img in IMAGES:
        ok = generate_image(img)
        results.append((img["filename"], ok))

    print(f"\n{'='*64}")
    print("Summary:")
    all_ok = True
    for fname, ok in results:
        full = os.path.join(OUTPUT_DIR, fname)
        if ok and os.path.exists(full):
            size = os.path.getsize(full)
            print(f"  [OK]     {fname}  ({size:,} bytes)")
        else:
            print(f"  [FAILED] {fname}")
            all_ok = False

    if not all_ok:
        sys.exit(1)
    else:
        print("\nAll 3 images generated successfully.")


if __name__ == "__main__":
    main()
