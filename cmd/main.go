package main

import (
	"kaiquecaires/real-time-leaderboard/cmd/auth"
	"kaiquecaires/real-time-leaderboard/cmd/db"
	"kaiquecaires/real-time-leaderboard/cmd/handlers"
	"kaiquecaires/real-time-leaderboard/cmd/messaging"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	route := gin.Default()
	route.GET("/", func(c *gin.Context) {
		c.IndentedJSON(http.StatusOK, "API IS ON FIRE!")
	})

	conn := db.GetPostgresInstance()
	userStore := db.NewPostgresUserStore(conn)
	gameStore := db.NewPostgresGameStore(conn)
	userScoreStore := db.NewPostgresUserScoreStore(conn)
	producer := messaging.GetProducer()
	userScorePublisher := messaging.NewKafkaUserScorePublisher(producer)

	signUpHandler := handlers.NewSignUpHandler(userStore)
	createGameHandler := handlers.NewGameHandler(gameStore)
	loginHandler := handlers.NewLoginHandler(userStore)
	userScoreHandler := handlers.NewUserScoreHandler(userScorePublisher, userScoreStore)

	route.POST("/login", loginHandler.Handle)
	route.POST("/signup", signUpHandler.Handle)

	authorized := route.Group("/", auth.AuthRequired)
	authorized.POST("/game", createGameHandler.CreateGameHandler)
	authorized.POST("/user-score", userScoreHandler.HandleSendUserScore)
	authorized.GET("/leaderboard", userScoreHandler.HandleGetLeaderboard)

	leaderboardConsumer := messaging.NewLeaderboardConsumer(userScoreStore)
	go leaderboardConsumer.Consume("leaderboard_postgres_1", "leaderdoard_postgres")

	route.Run("0.0.0.0:8080")
}
