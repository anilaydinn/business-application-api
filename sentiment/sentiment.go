package sentiment

import (
	"encoding/csv"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bbalet/stopwords"
	"github.com/euskadi31/go-tokenizer"
	"github.com/navossoc/bayesian"
)

type reviewData struct {
	comment string
	class   string
}

const (
	positive bayesian.Class = "positive"
	negative bayesian.Class = "negative"
)

func preProcessReviews(reviews []string) (reviewWords []string) { //reviewleri kelimelere ayırma
	re := regexp.MustCompile("[^a-zA-Z 0-9]+") //harf olmayanları sil
	t := tokenizer.New()
	for _, sentence := range reviews {
		newSentence := re.ReplaceAllString(strings.ToLower(sentence), "")  //harf olmayanları sil
		cleadnedSentence := stopwords.CleanString(newSentence, "en", true) //stopword sil
		tokenizedSentence := t.Tokenize(cleadnedSentence)
		for _, word := range tokenizedSentence {
			reviewWords = append(reviewWords, word)
		}
	}
	return reviewWords
}

func preProcessSentence(sentence string) (sentenceWords []string) { //sentenceleri kelimelere ayırarak tokenleştirme
	re := regexp.MustCompile("[^a-zA-Z 0-9]+") //harf olmayanları sil
	t := tokenizer.New()
	newSentence := re.ReplaceAllString(strings.ToLower(sentence), "")  //harf olmayanları sil
	cleadnedSentence := stopwords.CleanString(newSentence, "en", true) //stopword sil
	tokenizedSentence := t.Tokenize(cleadnedSentence)
	for _, word := range tokenizedSentence {
		sentenceWords = append(sentenceWords, word)
	}
	return sentenceWords
}

func main() {
	csvfile, err := os.Open("data/IMDBDataset.csv") //dosyayı al

	if err != nil {
		fmt.Println("csv açılamadi")
	}

	defer csvfile.Close() //program sonunda dosyayı kapa

	csvLines, err := csv.NewReader(csvfile).ReadAll() //dosyayı oku

	if err != nil {
		fmt.Println(err)
	}

	reviews := []reviewData{}

	for _, line := range csvLines {
		reviews = append(reviews, reviewData{
			comment: line[0],
			class:   line[1], //'1'. sutun
		})
	}

	positiveReview := []string{}
	negativeReview := []string{}

	for _, item := range reviews { //sadece reviewleri alma ve ayırma
		if item.class == "positive" {
			positiveReview = append(positiveReview, item.comment)
		}
		if item.class == "negative" {
			negativeReview = append(negativeReview, item.comment)
		}
	}
	positiveReviewWords := preProcessReviews(positiveReview)

	nagativeReviewWords := preProcessReviews(negativeReview)

	for _, item := range positiveReviewWords {
		fmt.Println(item)
	}
	for _, item := range nagativeReviewWords {
		fmt.Println(item)
	}

	classifier := bayesian.NewClassifier(positive, negative) //classları belirleme

	classifier.Learn(positiveReviewWords, positive) //classları atama
	classifier.Learn(nagativeReviewWords, negative)
	classifier.ConvertTermsFreqToTfIdf()

	comment := "Return to the 36th Chamber is one of those classic Kung-Fu movies which Shaw produces back in the 70s and 80s, whose genre is equivalent to the spaghetti westerns of Hollywood, and the protagonist Gordon Liu, the counterpart to the western's Clint Eastwood. Digitally remastered and a new print made for the Fantastic Film Fest"

	sentenceWords := preProcessSentence(comment)

	for _, item := range sentenceWords {
		fmt.Println(item)
	}

	fmt.Println(classifier.ProbScores(sentenceWords))

	_, result, _ := classifier.ProbScores(sentenceWords)

	fmt.Println(classifier.ProbScores(sentenceWords))

	if result == 0 {
		fmt.Println("positive")
	}
	if result == 1 {
		fmt.Println("negative")
	}

}
