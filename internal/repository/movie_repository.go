package repository

import (
	"context"
	"downloader_gochat/model"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"gorm.io/gorm"
)

type IMovieRepository interface {
	GetPosters(movieId string) ([]model.MoviePoster, error)
	SavePosters(movieId string, posters []model.MoviePoster) error
	SavePosterS3BlurHash(movieId string, posterUrl string, blurHash string) error
	SavePosterWideS3BlurHash(movieId string, posterUrl string, blurHash string) error
}

type MovieRepository struct {
	db      *gorm.DB
	mongodb *mongo.Database
}

func NewMovieRepository(db *gorm.DB, mongodb *mongo.Database) *MovieRepository {
	return &MovieRepository{db: db, mongodb: mongodb}
}

//------------------------------------------
//------------------------------------------

func (m *MovieRepository) GetPosters(movieId string) ([]model.MoviePoster, error) {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return nil, err
	}

	var result struct {
		Posters []model.MoviePoster `bson:"posters"`
	}
	opts := options.FindOne().SetProjection(bson.D{{"posters", 1}})
	err = m.mongodb.
		Collection("movies").
		FindOne(context.TODO(), bson.D{{"_id", id}}, opts).
		Decode(&result)
	return result.Posters, err
}

func (m *MovieRepository) SavePosters(movieId string, posters []model.MoviePoster) error {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return err
	}
	_, err = m.mongodb.
		Collection("movies").
		UpdateOne(context.TODO(),
			bson.D{{"_id", id}},
			bson.D{{"$set", bson.D{{"posters", posters}}}})
	return err
}

func (m *MovieRepository) SavePosterS3BlurHash(movieId string, posterUrl string, blurHash string) error {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return err
	}
	_, err = m.mongodb.
		Collection("movies").
		UpdateOne(context.TODO(),
			bson.D{{"_id", id}, {"poster_s3.url", posterUrl}},
			bson.D{{"$set", bson.D{{"poster_s3.blurHash", blurHash}}}})
	return err
}

func (m *MovieRepository) SavePosterWideS3BlurHash(movieId string, posterUrl string, blurHash string) error {
	id, err := primitive.ObjectIDFromHex(movieId)
	if err != nil {
		return err
	}
	_, err = m.mongodb.
		Collection("movies").
		UpdateOne(context.TODO(),
			bson.D{{"_id", id}, {"poster_wide_s3.url", posterUrl}},
			bson.D{{"$set", bson.D{{"poster_wide_s3.blurHash", blurHash}}}})
	return err
}
