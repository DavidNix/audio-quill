# Audio Quill

Transcribe WAV files to text files with titles using Whisper and Ollama.

All processing is local. No data leaves your machine.

### Background:

I carry a low-tech ~$20 audio recorder when I go on walks to capture my thoughts and ideas. I leave my phone at home.

That way, I'm not distracted or tempted to check email, slack, discord, social media, etc.

Then when ready, I plug the SD card into my computer and transcribe the files.

I've been very happy [with this device](https://www.amazon.com/dp/B0CKRBSM1X?psc=1&ref=ppx_yo2ov_dt_b_product_details).

But audio isn't as useful as text. So this script turns them into text which I can organize later.

## Setup

Download [whisperfile](https://huggingface.co/Mozilla/whisperfile).

```sh
wget https://huggingface.co/Mozilla/whisperfile/resolve/main/whisper-tiny.en.llamafile
chmod +x whisper-tiny.en.llamafile
```

Make sure [Ollama](https://ollama.com/) is running and you've downloaded llama3.1 model.

```sh
ollama run llama3.1
```

## Usage

IMPORTANT: I've only tested this on my M2 MacBook. YMMV.

Run the app

```sh
go run . --source /Volumes/USB-DISK/RECORD --dest ~/Downloads/aquill1
```
