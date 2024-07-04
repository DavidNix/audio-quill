import os
import argparse
import whisper
import openai


def generate_title(transcription):
    prompt = f"Generate a concise and relevant title for the following transcription:\n\n{transcription}\n\nTitle:"
    response = openai.Completion.create(
        engine="gpt-3.5-turbo",
        prompt=prompt,
        max_tokens=50,
        n=1,
        stop=None,
        temperature=0.7,
    )
    title = response.choices[0].text.strip()
    return title


def transcribe_audio_files(source_folder, destination_folder):
    model = whisper.load_model("base")

    for filename in os.listdir(source_folder):
        if filename.endswith(".wav"):
            print(f"Transcribing: {filename}")
            file_path = os.path.join(source_folder, filename)

            result = model.transcribe(file_path)

            transcription = result["text"]

            title = generate_title(transcription)
            print(f"Saving: {title}.md")

            markdown_path = os.path.join(destination_folder, title.replace(" ", "_") + ".md")

            with open(markdown_path, "w") as file:
                file.write(f"# {title}\n\n{transcription}")

            print(f"Transcription saved: {markdown_path}")


# Set up OpenAI API credentials
openai.api_key = "your_openai_api_key"

# Create an argument parser
parser = argparse.ArgumentParser(description="Transcribe audio files and generate markdown files.")
parser.add_argument("source_folder", help="Path to the folder containing the .WAV files.")
parser.add_argument("destination_folder", help="Path to the folder where the markdown files will be saved.")

# Parse the command line arguments
args = parser.parse_args()

# Call the function to transcribe the audio files
transcribe_audio_files(args.source_folder, args.destination_folder)
