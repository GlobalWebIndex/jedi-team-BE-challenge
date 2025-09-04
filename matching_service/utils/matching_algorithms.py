from question_sanitiser import remove_stopwords
from sklearn.feature_extraction.text import ENGLISH_STOP_WORDS
from rapidfuzz import fuzz
import re

def most_common_word_match(query, sentences):

    query_words = set(remove_stopwords(query))
    max_score = 0
    best_match = None

    for sentence in sentences:

        sentence_words = set(remove_stopwords(sentence))
        common = query_words & sentence_words
        score = len(common)

        if score > max_score:
            max_score = score
            best_match = sentence

    return best_match, max_score

def tokenize(text):
    # Clean punctuation, lowercase, remove stopwords
    text = re.sub(r'[^\w\s]', '', text.lower())
    return [word for word in text.split() if word not in ENGLISH_STOP_WORDS]

def common_fuzzy_word_count(s1, s2, threshold=3):
    words1 = tokenize(s1)
    words2 = tokenize(s2)
    count = 0
    for w1 in words1:
        for w2 in words2:
            if fuzz.ratio(w1, w2) >= threshold:
                count += 1
                break  # only count once per word
    return count

def best_fuzzy_match(query, sentences, threshold=3):
    max_score = 0
    best_sentence = None

    for sentence in sentences:
        score = common_fuzzy_word_count(query, sentence, threshold)
        # score = fuzz.token_set_ratio(query, sentence)

        if score > max_score and score >=threshold:
            max_score = score
            best_sentence = sentence

    return best_sentence, max_score
