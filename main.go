package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/h2non/bimg"
	"github.com/joho/godotenv"
)

var invalid bool = false

type File struct {
	OK     bool
	Result struct {
		FileID       string `json:"file_id"`
		FileUniqueID string `json:"file_unique_id"`
		FileSize     int    `json:"file_size,omitempty"`
		FilePath     string `json:"file_path,omitempty"`
	} `json:"result"`
}

func main() {

	err := godotenv.Load("api.env")
	if err != nil {
		log.Fatalf("Some error occured. Err: %s", err)
	}

	API_KEY := os.Getenv("APIKEY")

	bot, err := tgbotapi.NewBotAPI(API_KEY)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	//dislpay username and authorized account
	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message.Text == "/start" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Hi!! 😁 \nWelcome to Imagecompressorbot\nSend Documented image only don't compress image while sending \nSelect your image ratio portrait or landscape \nyou will receive your compressed image \nThank you! 😄")
			bot.Send(msg)
		}
		//Only allow documented image and text
		if (update.Message.Document != nil) || (update.Message.Text != "") { // If we got a message
			log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			log.Printf("<------------->")
			st := update.Message.Text

			log.Printf((st))
			//Just to take documneted image as input and avoid error
			if st == "" {
				//TODO: input custom resize value
				Fid := update.Message.Document.FileID //File id
				Cid := update.Message.Chat.ID         //chat id
				portrait := false
				landscape := false

				log.Printf("%s", Fid)
				bot.GetFile(tgbotapi.FileConfig{Fid})
				urlgetpath := "https://api.telegram.org/bot" + API_KEY + "/getFile?file_id=" + Fid
				Fpath := get_content(urlgetpath)

				urldownload := "https://api.telegram.org/file/bot" + API_KEY + "/" + Fpath
				go downloadFromUrl(urldownload)

				msg1 := tgbotapi.NewMessage(Cid, "Documnented Image File Received")
				msg1.ReplyToMessageID = update.Message.MessageID
				bot.Send(msg1)

				msg2 := tgbotapi.NewMessage(Cid, "Processing...")
				bot.Send(msg2)

				msg := tgbotapi.NewMessage(Cid, "Send your photo ratio 'Landscape' or 'Portrait'\nExample if your image is Portrait send message Portrait")
				bot.Send(msg)

				//wait for userinput
				for update := range updates {

					if update.Message.Text == "Portrait" {
						log.Printf("portrait")
						portrait = true
						break
					} else if update.Message.Text == "Landscape" {
						log.Printf("landscape")
						landscape = true
						break
					} else {
						msg := tgbotapi.NewMessage(Cid, "Please send Landscape or Portrait ,Check your spelling")
						bot.Send(msg)
					}
				}

				if portrait {
					p := compress(Fpath, 720, 1280)

					if invalid {
						msginvalid := tgbotapi.NewMessage(Cid, "Invalid file input")
						bot.Send(msginvalid)
					} else {
						msg3 := tgbotapi.NewDocument(Cid, tgbotapi.FilePath("output/"+p))
						bot.Send(msg3)
					}
				} else if landscape {
					p := compress(Fpath, 1280, 720)

					if invalid {
						msginvalid := tgbotapi.NewMessage(Cid, "Invalid file input")
						bot.Send(msginvalid)
					} else {
						msg3 := tgbotapi.NewDocument(Cid, tgbotapi.FilePath("output/"+p))
						bot.Send(msg3)
					}

				} else {
					log.Printf("=====================")
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please check spelling")
					msg.ReplyToMessageID = update.Message.MessageID
					bot.Send(msg)
				}

			} else {
				if update.Message.Text == "/start" {

				} else {

					Cid := update.Message.Chat.ID
					msgerr := tgbotapi.NewMessage(Cid, "First send documneted image")
					bot.Send(msgerr)
				}

			}
		} else if update.Message.Photo != nil {
			log.Print("Photo is compressed")
			msgphoto := tgbotapi.NewMessage(update.Message.Chat.ID, "Don't send compressed image only send documented image")
			bot.Send(msgphoto)
		} else {
			if update.Message.Text == "/start" {

			} else {
				Cid := update.Message.Chat.ID

				msgerr := tgbotapi.NewMessage(Cid, "please Send documneted image and don't compress image")
				bot.Send(msgerr)
			}
		}

	}
}

// download file to database
func downloadFromUrl(url string) {
	tokens := strings.Split(url, "/")
	fileName := "documents/" + tokens[len(tokens)-1]
	fmt.Println("Downloading", url, "to", fileName)

	output, err := os.Create(fileName)
	if err != nil {
		fmt.Println("Error while creating", fileName, "-", err)
		return
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		fmt.Println("Error while downloading", url, "-", err)
		return
	}

	fmt.Println(n, "bytes downloaded.")
}

// get json response from api
func get_content(upath string) string {
	// json data
	url := upath

	res, err := http.Get(url)

	if err != nil {
		panic(err.Error())
	}

	body, err := ioutil.ReadAll(res.Body)

	if err != nil {
		panic(err.Error())
	}

	var data File
	json.Unmarshal(body, &data)
	Fpath := data.Result.FilePath
	fmt.Printf("Results: %v\n", data.Result.FilePath)
	return Fpath
	//os.Exit(0)
}

// compress image
func compress(Fpath string, width int, height int) string {

	buffer, err := bimg.Read(Fpath)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	newImage, err := bimg.NewImage(buffer).Resize(width, height)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		invalid = true
	}
	//
	size, err := bimg.NewImage(newImage).Size()
	if size.Width == width && size.Height == height {
		fmt.Println("The image size is valid")
		invalid = false
	}

	splitpath := strings.Split(Fpath, "/")
	p := splitpath[1]

	bimg.Write("output/"+p, newImage)
	log.Printf("output/" + p)

	return p
}
