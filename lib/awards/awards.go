package awards

import (
	"confesi/config"
	"confesi/db"

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
