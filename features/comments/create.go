package comments

import (
	"confesi/db"
	"confesi/lib/response"
	"confesi/lib/utils"
	"confesi/lib/validation"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// func doesIdentifierExist(tx *gorm.DB, UID string, postId uint, identifier *uint) (error, db.CommentIdentifier, bool) {
// 	// check if user has already commented on this post with the same matchings
// 	possilbeIdentifier := db.CommentIdentifier{}

// 	query := tx.
// 		Where("user_id = ?", UID).
// 		Where("post_id = ?", postId).
// 		Where("identifier = ?", identifier)

// 	err := query.First(&possilbeIdentifier).Error
// 	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 		return serverError, possilbeIdentifier, false
// 	} else if errors.Is(err, gorm.ErrRecordNotFound) {
// 		// we have to create it and link the comment to this new one's ID
// 		return nil, possilbeIdentifier, false
// 	} else {
// 		// we link the comment to the ID of the existing one
// 		return nil, possilbeIdentifier, true
// 	}
// }

// (error, bool, uint) -> (error, alreadyPosted, numericalUser)
func getAlreadyPostedNumericalUser(tx *gorm.DB, postID uint, userID string) (error, bool, uint) {
	comment := db.Comment{}
	err := tx.
		Where("user_id = ?", userID).
		Where("post_id = ?", postID).
		Where("numerical_user IS NOT NULL").
		First(&comment).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return serverError, false, 0
	} else if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, false, 0
	} else {
		return nil, true, *comment.Numerics.NumericalUser
	}
}

func getNextIdentifier(tx *gorm.DB, postId uint) (error, uint) {
	highestIdentifier := db.Comment{}
	err := tx.
		Where("post_id = ?", postId).
		Order("numerical_user ASC").
		Find(&highestIdentifier).
		Limit(1).
		Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		fmt.Println(err, "HERE 1")
		return serverError, 0
	}
	if errors.Is(err, gorm.ErrRecordNotFound) || highestIdentifier.Numerics.NumericalUser == nil {
		fmt.Println(err, "HERE 2")
		return nil, 1
	} else {
		fmt.Println(err, "HERE 3")
		return nil, *highestIdentifier.Numerics.NumericalUser + 1
	}
}

func (h *handler) handleCreate(c *gin.Context) {

	// validate the json body from request
	var req validation.CreateComment
	err := utils.New(c).Validate(&req)
	if err != nil {
		return
	}

	// get user token
	token, err := utils.UserTokenFromContext(c)
	if err != nil {
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// start a transaction
	tx := h.db.Begin()

	// if something goes ary, rollback
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err("server error").Send(c)
			return
		}
	}()

	var post db.Post
	err = tx.
		Where("id = ?", req.PostID).
		First(&post).
		Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.New(http.StatusBadRequest).Err("post not found").Send(c)
		}
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		tx.Rollback()
		return
	}

	fmt.Println("GOT HEREEEE1")

	isOp := post.UserID == token.UID

	fmt.Println(post.UserID, token.UID)

	fmt.Println("OP CHECK IS ", isOp)

	fmt.Println("GOT HEREEEE2")

	// base comment
	comment := db.Comment{
		UserID:  token.UID,
		PostID:  req.PostID,
		Content: req.Content,
	}

	parentComment := db.Comment{}

	// they are trying to create a threaded comment
	if req.ParentCommentID != nil {

		// parent comment

		err = tx.
			Where("comments.id = ?", req.ParentCommentID).
			Where("comments.post_id = ?", req.PostID).
			Find(&parentComment).
			Updates(map[string]interface{}{
				"children_count": gorm.Expr("children_count + ?", 1),
			}).
			Error
		if err != nil {
			// parent comment not found
			if errors.Is(err, gorm.ErrRecordNotFound) {
				tx.Rollback()
				response.New(http.StatusBadRequest).Err("parent-comment and post combo doesn't exist").Send(c)
				return
			}
			// some other error
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		if parentComment.ParentRoot == nil {
			comment.ParentRoot = &parentComment.ID
		} else {
			comment.ParentRoot = parentComment.ParentRoot
		}
	} else {
		comment.ParentRoot = nil
	}

	var nextIdentifier uint
	// is OP?
	if !isOp {
		err, nextIdentifier = getNextIdentifier(tx, req.PostID)
		if err != nil {
			fmt.Println("GOT HEREEEE4")
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
	}

	fmt.Println("GOT HEREEEE3")

	if parentComment.Numerics.NumericalUserIsOp {
		comment.Numerics.NumericalReplyingUserIsOp = true
	} else {
		comment.Numerics.NumericalReplyingUserIsOp = false
		comment.Numerics.NumericalReplyingUser = parentComment.Numerics.NumericalUser
	}

	if isOp {
		fmt.Println("FOUND THE OPPPPPPPP HCEK PASSED")
		comment.Numerics.NumericalUserIsOp = true
	} else {
		comment.Numerics.NumericalUserIsOp = false
		err, alreadyPosted, userNumeric := getAlreadyPostedNumericalUser(tx, req.PostID, token.UID)
		if err != nil {
			tx.Rollback()
			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
			return
		}
		if alreadyPosted {
			comment.Numerics.NumericalUser = &userNumeric
		} else {
			comment.Numerics.NumericalUser = &nextIdentifier
		}

	}

	// create the comment
	err = tx.Create(&comment).
		Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}

	// if all goes well, respond with a 201 & commit the transaction
	err = tx.Commit().Error
	if err != nil {
		tx.Rollback()
		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
		return
	}
	response.New(http.StatusCreated).Send(c)
}
