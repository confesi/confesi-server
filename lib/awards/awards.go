package awards

import (
	"confesi/db"

	"gorm.io/gorm"
)

func OnPostVote(tx *gorm.DB, upvotes uint, downvotes uint, contentID db.EncryptedID) error {
	// Define a struct that represents the structure of your post/comment
	var content struct {
		UserID string `gorm:"column:user_id"`
	}

	// Retrieve the user_id of the content's owner
	err := tx.Model(&db.Post{}).Select("user_id").Where("id = ?", contentID).Scan(&content).Error
	if err != nil {
		tx.Rollback()
		return err
	}

	// Check if the post/comment has exactly 100 net upvotes (upvotes - downvotes)
	if upvotes-downvotes == 100 {
		var award db.Award
		whereAward := db.Award{
			AwardID:   db.EncryptedID{Val: 1}, // Set the actual award id
			PostID:    &contentID,             // Assuming this is a post
			CommentID: nil,                    // or set this if it's a comment
			UserID:    content.UserID,
		}

		// Check if an award already exists, if not create a new one
		result := tx.Where(&whereAward).FirstOrCreate(&award)
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}

		// If the award already existed, increment the quantity
		if result.RowsAffected == 0 {
			err := tx.Model(&award).Update("quantity", gorm.Expr("quantity + ?", 1)).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// If everything went well, the transaction should be committed outside of this function
	return nil
}

func OnCommentVote(db *gorm.DB, upvotes uint, downvotes uint, commentId db.EncryptedID) error {
	return nil
}

func OnPostBecomingHottest(db *gorm.DB, post db.Post) error {
	return nil
}
