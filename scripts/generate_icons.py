"""
generate_icons.py — Generates multi-resolution Windows ICO and web icon assets
from MobRig brand logo files in assets/brand/.
"""

import os
import sys
from PIL import Image

REPO_ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), ".."))
BRAND_DIR = os.path.join(REPO_ROOT, "assets", "brand")
ICONS_DIR = os.path.join(REPO_ROOT, "assets", "icons")
WINRES_DIR = os.path.join(REPO_ROOT, "cmd", "mobrig", "winres")
UI_PUBLIC_DIR = os.path.join(REPO_ROOT, "ui", "public")

# Standard Windows icon sizes (16px to 256px)
ICON_SIZES = [(16, 16), (24, 24), (32, 32), (48, 48), (64, 64), (128, 128), (256, 256)]


def clean_image(src_path: str, is_transparent: bool = True) -> Image.Image:
    im = Image.open(src_path)
    if is_transparent:
        im = im.convert("RGBA")
        r, g, b, a = im.split()
        # Clean subtle compression artifacts (alpha <= 2 becomes 0)
        a = a.point(lambda p: 0 if p <= 2 else p)
        return Image.merge("RGBA", (r, g, b, a))
    return im.convert("RGB")


def create_ico(clean_im: Image.Image, out_path: str) -> None:
    os.makedirs(os.path.dirname(out_path), exist_ok=True)
    clean_im.save(out_path, format="ICO", sizes=ICON_SIZES)
    file_size = os.path.getsize(out_path)
    with Image.open(out_path) as ico:
        sizes_found = ico.info.get("sizes", set())
    print(f"  -> {os.path.relpath(out_path, REPO_ROOT)}: {file_size} bytes, sizes: {sorted(sizes_found)}")


def main():
    print(f"Generating MobRig icons from assets in: {BRAND_DIR}")

    dark_logo = os.path.join(BRAND_DIR, "mobrig-logo-dark.png")
    light_logo = os.path.join(BRAND_DIR, "mobrig-logo-light.png")
    brand_logo = os.path.join(BRAND_DIR, "mobrig-logo.png")

    for f in [dark_logo, light_logo, brand_logo]:
        if not os.path.exists(f):
            print(f"Error: required source logo not found: {f}", file=sys.stderr)
            sys.exit(1)

    print("\n[1/3] Processing Dark Variant (mobrig-logo-dark.png)...")
    im_dark = clean_image(dark_logo, is_transparent=True)
    create_ico(im_dark, os.path.join(ICONS_DIR, "mobrig.ico"))
    create_ico(im_dark, os.path.join(ICONS_DIR, "mobrig-dark.ico"))
    create_ico(im_dark, os.path.join(WINRES_DIR, "icon.ico"))
    create_ico(im_dark, os.path.join(WINRES_DIR, "icon-dark.ico"))

    print("\n[2/3] Processing Light Variant (mobrig-logo-light.png)...")
    im_light = clean_image(light_logo, is_transparent=True)
    create_ico(im_light, os.path.join(ICONS_DIR, "mobrig-light.ico"))
    create_ico(im_light, os.path.join(WINRES_DIR, "icon-light.ico"))
    create_ico(im_light, os.path.join(UI_PUBLIC_DIR, "favicon.ico"))

    print("\n[3/3] Processing Brand Logo (mobrig-logo.png)...")
    im_brand = clean_image(brand_logo, is_transparent=False)
    create_ico(im_brand, os.path.join(ICONS_DIR, "mobrig-logo.ico"))
    create_ico(im_brand, os.path.join(WINRES_DIR, "icon-logo.ico"))

    print("\nAll icon assets successfully generated.")


if __name__ == "__main__":
    main()
