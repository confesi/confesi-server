package awards

import (
	"confesi/config"
	"confesi/db"
	"errors"
	"fmt"
	"regexp"

	"gorm.io/gorm"
)

// Award for a user getting more than X upvotes on a post
func AwardPostGreaterThanXUpvotes(tx *gorm.DB, upvotes uint, downvotes uint, postID db.EncryptedID) error {
	awardTypeID := db.EncryptedID{Val: config.AwardTypesPostGreaterThanXUpvotes}

	if upvotes-downvotes == 100 {
		// Retrieve the user_id of the content's owner
		var content struct {
			UserID string
		}
		err := tx.Model(&db.Post{}).Select("user_id").Where("id = ?", postID).Scan(&content).Error
		if err != nil {
			tx.Rollback()
			return err
		}
		// Check if an award has already been given for this content and user
		var existingAward db.AwardsGeneral
		result := tx.Where("post_id = ? AND user_id = ? AND award_type_id = ?", postID, content.UserID, awardTypeID).First(&existingAward)
		if result.RowsAffected == 0 {
			// No existing award found, create new award entry
			awardGeneral := db.AwardsGeneral{
				PostID:      &postID, // Assuming this is a post. For comments, adjust accordingly.
				CommentID:   nil,     // Set this if it's a comment.
				UserID:      content.UserID,
				AwardTypeID: awardTypeID,
			}

			err = tx.Create(&awardGeneral).Error
			if err != nil {
				tx.Rollback()
				return err
			}

			// Since a new award is created, update total in awards_total
			var totalAward db.AwardsTotal
			err = tx.Where(db.AwardsTotal{UserID: content.UserID, AwardTypeID: awardTypeID}).FirstOrCreate(&totalAward).Error
			if err != nil {
				tx.Rollback()
				return err
			}

			err = tx.Model(&totalAward).Update("total", gorm.Expr("total + ?", 1)).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

// Award for a user getting more than X upvotes on a comment
func AwardCommentGreaterThanXUpvotes(tx *gorm.DB, upvotes uint, downvotes uint, commentID db.EncryptedID) error {
	awardTypeID := db.EncryptedID{Val: config.AwardTypesCommentGreaterThanXUpvotes}

	if upvotes-downvotes == 50 {
		// Retrieve the user_id of the content's owner
		var comment struct {
			UserID string
		}
		err := tx.Model(&db.Comment{}).Select("user_id").Where("id = ?", commentID).Scan(&comment).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// Check if an award has already been given for this comment and user
		var existingAward db.AwardsGeneral
		result := tx.Where("comment_id = ? AND user_id = ? AND award_type_id = ?", commentID, comment.UserID, awardTypeID).First(&existingAward)
		if result.RowsAffected == 0 {
			// No existing award found, create new award entry
			awardGeneral := db.AwardsGeneral{
				PostID:      nil,        // Set this if it's a post.
				CommentID:   &commentID, // Assuming this is a comment
				UserID:      comment.UserID,
				AwardTypeID: awardTypeID,
			}

			err = tx.Create(&awardGeneral).Error
			if err != nil {
				tx.Rollback()
				return err
			}

			// Since a new award is created, update total in awards_total
			var totalAward db.AwardsTotal
			err = tx.Where(db.AwardsTotal{UserID: comment.UserID, AwardTypeID: awardTypeID}).FirstOrCreate(&totalAward).Error
			if err != nil {
				tx.Rollback()
				return err
			}

			err = tx.Model(&totalAward).Update("total", gorm.Expr("total + ?", 1)).Error
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return nil
}

func OnPostCreation(tx *gorm.DB, title string, body string, postID db.EncryptedID, userID string) error {
	// Expressive: Awarded for posting a confession with more than 10 different emojis
	// Expressionless: Awarded for posting a confession with more than 5000 characters with no emojis
	emojiRx := regexp.MustCompile(config.EmojiRegex)

	emojis := emojiRx.FindAllString(body+title, -1)
	uniqueEmojis := make(map[string]bool)

	for _, emoji := range emojis {
		uniqueEmojis[emoji] = true
	}

	// For Expressive award
	if len(uniqueEmojis) > 10 {
		expressiveAwardID := db.EncryptedID{Val: config.AwardTypesLotsOfEmojis}
		if err := grantAward(tx, userID, postID, expressiveAwardID); err != nil {
			return err
		}
	}

	// For Expressionless award
	if len(body+title) > 5000 && len(emojis) == 0 {
		expressionlessAwardID := db.EncryptedID{Val: config.AwardTypesNoEmojisLargePost}
		if err := grantAward(tx, userID, postID, expressionlessAwardID); err != nil {
			return err
		}
	}

	return nil
}

func grantAward(tx *gorm.DB, userID string, postID db.EncryptedID, awardTypeID db.EncryptedID) error {
	fmt.Println("grantAward: ", postID, userID, awardTypeID)

	// Check for an existing award
	var existingAward db.AwardsGeneral
	result := tx.Where("post_id = ? AND user_id = ? AND award_type_id = ?", postID, userID, awardTypeID).First(&existingAward)
	if result.RowsAffected == 0 {
		// Create a new award entry
		newAward := db.AwardsGeneral{
			PostID:      &postID,
			UserID:      userID,
			AwardTypeID: awardTypeID,
		}

		if err := tx.Create(&newAward).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	// Update or create total awards count for the user
	var totalAward db.AwardsTotal
	err := tx.Where("user_id = ? AND award_type_id = ?", userID, awardTypeID).First(&totalAward).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new total if not found
			totalAward = db.AwardsTotal{
				UserID:      userID,
				AwardTypeID: awardTypeID,
				Total:       1,
			}
			if err = tx.Create(&totalAward).Error; err != nil {
				tx.Rollback()
				return err
			}
		} else {
			tx.Rollback()
			return err
		}
	} else {
		// Update existing total
		if err = tx.Model(&totalAward).Update("total", gorm.Expr("total + ?", 1)).Error; err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}

func OnPostBecomingHottest(tx *gorm.DB, postIDs []db.EncryptedID) error {
	hottestPostAwardTypeID := db.EncryptedID{Val: config.AwardTypesPostBecomingHottest}

	for _, postID := range postIDs {
		// Retrieve the user_id of the post's owner
		var post struct {
			UserID string
		}
		err := tx.Model(&db.Post{}).Select("user_id").Where("id = ?", postID).Scan(&post).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// Insert a new entry in awards_general
		awardGeneral := db.AwardsGeneral{
			PostID:      &postID,
			CommentID:   nil,
			UserID:      post.UserID,
			AwardTypeID: hottestPostAwardTypeID,
		}

		err = tx.Create(&awardGeneral).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		// Update total in awards_total
		var totalAward db.AwardsTotal
		err = tx.Where(db.AwardsTotal{UserID: post.UserID, AwardTypeID: hottestPostAwardTypeID}).FirstOrCreate(&totalAward).Error
		if err != nil {
			tx.Rollback()
			return err
		}

		err = tx.Model(&totalAward).Update("total", gorm.Expr("total + ?", 1)).Error
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return nil
}
