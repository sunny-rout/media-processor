import yt_dlp
import json
import subprocess

def download_instagram_image(url: str, output_path: str = "-") -> bytes:                                 
    """                                                                                                        Download an image from an Instagram URL using gallery-dl.                                            

    Args:
        url: Instagram post URL (e.g., https://www.instagram.com/p/ABC123/)
        output_path: Output path. Use "-" for stdout (pipe mode).

    Returns:
        The downloaded image data as bytes (if output_path is "-").
    """
    result = subprocess.run(
        ["gallery-dl", "--cookies-from-browser", "chrome", "-o", output_path, url],
        capture_output=True,
        text=True,
    )

    if result.returncode != 0:
        raise RuntimeError(f"gallery-dl failed: {result.stderr}")

    return result.stdout

def download_instagram(url):
    ydl_opts = {
        'quiet': False,
        'outtmpl': '%(title)s_%(id)s.%(ext)s',
        'format': 'best',  # IMPORTANT: allow non-video formats
        'ignoreerrors': True,  # prevents crash on non-video
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        info = ydl.extract_info(url, download=True)

        # Handle carousel (multiple images/videos)
        if info.get("_type") == "playlist":
            print("\n📚 Carousel detected")
            for entry in info.get("entries", []):
                print("Downloaded:", entry.get("url"))
        else:
            print("\n📸 Single media")
            print("Downloaded:", info.get("url"))

def download_instagram_media(url):
    ydl_opts = {
        'outtmpl': '%(title)s.%(ext)s',  # output filename
        'quiet': False,
    }

    with yt_dlp.YoutubeDL(ydl_opts) as ydl:
        info = ydl.extract_info(url, download=True)

        print("\n✅ Download complete!")
        print("Title:", info.get("title"))
        print("Type:", info.get("_type"))
        print("Formats available:", len(info.get("formats", [])))

def inspect_instagram(url):
    with yt_dlp.YoutubeDL({'quiet': True}) as ydl:
        info = ydl.extract_info(url, download=False)

        print(json.dumps(info, indent=2))

def stream_instagram(url):
    command = [
        "yt-dlp",
        "-o", "-",   # stream to stdout
        url
    ]

    process = subprocess.Popen(command, stdout=subprocess.PIPE)

    with open("output.bin", "wb") as f:
        while True:
            chunk = process.stdout.read(1024 * 32)
            if not chunk:
                break
            f.write(chunk)

    print("✅ Streamed to output.bin")

if __name__ == "__main__":
    url = input("Enter Instagram URL: ")
    download_instagram_image(url)