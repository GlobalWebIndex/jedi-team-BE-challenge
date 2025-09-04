from sklearn.feature_extraction.text import ENGLISH_STOP_WORDS
import ollama

def remove_stopwords(text):
    words = text.lower().split()
    filtered = [word for word in words if word not in ENGLISH_STOP_WORDS]
    return filtered