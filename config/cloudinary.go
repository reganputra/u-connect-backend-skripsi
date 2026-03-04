package config

import (
	"log"
	"os"

	"github.com/cloudinary/cloudinary-go/v2"
)

var Cloudinary *cloudinary.Cloudinary

func ConnectCloudinary() {
	cloudinaryURL := os.Getenv("CLOUDINARY_URL")
	if cloudinaryURL == "" {
		log.Fatal("❌ CLOUDINARY_URL is not set in environment")
	}

	cld, err := cloudinary.NewFromURL(cloudinaryURL)
	if err != nil {
		log.Fatalf("❌ Failed to initialize Cloudinary: %v", err)
	}

	Cloudinary = cld
	log.Println("✅ Cloudinary initialized successfully!")
}
