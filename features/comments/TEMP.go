// package comments

// import (
// 	"confesi/config"
// 	"confesi/db"
// 	"confesi/lib/response"
// 	"confesi/lib/utils"
// 	"confesi/lib/validation"
// 	"errors"
// 	"fmt"
// 	"net/http"

// 	"github.com/gin-gonic/gin"
// 	"github.com/lib/pq"
// 	"gorm.io/gorm"
// )

// func (h *handler) handleCreate(c *gin.Context) {

// 	// validate the json body from request
// 	var req validation.CreateComment
// 	err := utils.New(c).Validate(&req)
// 	if err != nil {
// 		return
// 	}

// 	// get user token
// 	token, err := utils.UserTokenFromContext(c)
// 	if err != nil {
// 		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 		return
// 	}

// 	// start a transaction
// 	tx := h.db.Begin()

// 	// if something goes ary, rollback
// 	defer func() {
// 		if r := recover(); r != nil {
// 			fmt.Println("ROLL BACK")

// 			tx.Rollback()
// 			response.New(http.StatusInternalServerError).Err("server error").Send(c)
// 			return
// 		}
// 	}()

// 	// base comment
// 	comment := db.Comment{
// 		UserID:  token.UID,
// 		PostID:  req.PostID,
// 		Content: req.Content,
// 	}

// 	futureParentIdentifier := db.CommentIdentifier{}
// 	parentComment := db.Comment{}
// 	// they are trying to create a threaded comment
// 	if req.ParentCommentID != nil {
// 		err = tx.
// 			Model(&parentComment).
// 			Joins("JOIN comment_identifiers ON comments.identifier_id = comment_identifiers.id").
// 			Where("comments.id = ?", req.ParentCommentID).
// 			Where("comments.post_id = ?", req.PostID).
// 			UpdateColumns(map[string]interface{}{
// 				"children_count": gorm.Expr("children_count + ?", 1),
// 			}).
// 			First(&futureParentIdentifier).
// 			Error
// 		if err != nil {
// 			// parent comment not found
// 			if errors.Is(err, gorm.ErrRecordNotFound) {
// 				fmt.Println("ROLL BACK")

// 				tx.Rollback()
// 				response.New(http.StatusBadRequest).Err("parent-comment and post combo doesn't exist").Send(c)
// 				return
// 			}
// 			// some other error
// 			fmt.Println("ROLL BACK")

// 			tx.Rollback()
// 			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 			return
// 		}
// 		if len(parentComment.Ancestors) > config.MaxCommentThreadDepthExcludingRoot-1 {
// 			// can't thread comments this deep
// 			fmt.Println("ROLL BACK")

// 			tx.Rollback()
// 			response.New(http.StatusBadRequest).Err(threadDepthError.Error()).Send(c)
// 			return
// 		}
// 		comment.Ancestors = append(parentComment.Ancestors, *req.ParentCommentID)
// 	} else {
// 		// its a root comment
// 		comment.Ancestors = pq.Int64Array{}
// 	}

// 	// try to create a new identifier record
// 	var post db.Post
// 	err = tx.
// 		Where("id = ?", req.PostID).
// 		First(&post).
// 		Error
// 	if err != nil {
// 		if errors.Is(err, gorm.ErrRecordNotFound) {
// 			response.New(http.StatusBadRequest).Err("referenced post not found").Send(c)
// 		}
// 		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 		fmt.Println("ROLL BACK")

// 		tx.Rollback()
// 		return
// 	}
// 	// is the user the OP?
// 	if post.UserID == token.UID {
// 		// user is OP
// 		newOpCommentIdentifier := db.CommentIdentifier{
// 			UserID: token.UID,
// 			PostID: req.PostID,
// 			IsOp:   true,
// 		}
// 		// they're creating a threaded comment
// 		if req.ParentCommentID != nil {
// 			newOpCommentIdentifier.ParentIdentifier = futureParentIdentifier.Identifier
// 		}
// 		err = tx.Create(&newOpCommentIdentifier).Error
// 		if err != nil {
// 			tx.Rollback()
// 			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 			return
// 		}
// 		comment.IdentifierID = newOpCommentIdentifier.ID
// 	} else {
// 		// user is not OP
// 		var highestIdentifierSoFar db.CommentIdentifier
// 		var newIdentifier uint

// 		// check if there already exists a comment identifier fir user_id and post_id combo
// 		alreadyExistingCommentIdentifier := db.CommentIdentifier{}
// 		err = tx.
// 			Where("user_id = ?", token.UID).
// 			Where("post_id = ?", req.PostID).
// 			First(&alreadyExistingCommentIdentifier).
// 			Error
// 		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 			fmt.Println("ROLL BACK")

// 			tx.Rollback()
// 			response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 			return
// 		} else if errors.Is(err, gorm.ErrRecordNotFound) {
// 			// nothing found! the user has yet to comment on this post
// 			// list all the already existing comment identifiers and get the one with the highest "identifier" column, then save one with that + 1
// 			err = tx.
// 				Where("post_id = ?", req.PostID).
// 				Order("identifier ASC").
// 				Find(&highestIdentifierSoFar).
// 				Limit(1).
// 				Error
// 			fmt.Println("POINT 0")
// 			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 				fmt.Println("ROLL BACK")

// 				tx.Rollback()
// 				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 				return
// 			}
// 			fmt.Println("POINT 1")
// 			if errors.Is(err, gorm.ErrRecordNotFound) || highestIdentifierSoFar.Identifier == nil {
// 				newIdentifier = 1
// 			} else {
// 				newIdentifier = *highestIdentifierSoFar.Identifier + 1
// 			}
// 			fmt.Println("POINT 2")
// 			// save new comment identifier
// 			newNotOpCommentIdentifier := db.CommentIdentifier{
// 				UserID:     token.UID,
// 				PostID:     req.PostID,
// 				Identifier: &newIdentifier,
// 				IsOp:       false,
// 			}
// 			fmt.Println("POINT 3")
// 			// they're creating a threaded comment // todo
// 			if req.ParentCommentID != nil {
// 				newNotOpCommentIdentifier.ParentIdentifier = futureParentIdentifier.Identifier
// 			}
// 			fmt.Println("POINT 4")
// 			err = tx.Create(&newNotOpCommentIdentifier).
// 				Error
// 			if err != nil {
// 				fmt.Println("ROLL BACK")

// 				tx.Rollback()
// 				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 				return
// 			}
// 			fmt.Println("POINT 5")
// 			comment.IdentifierID = newNotOpCommentIdentifier.ID
// 		} else {
// 			fmt.Println("POINT 6")
// 			// todo: use the existing one, but with new parent_identifier
// 			resaveCommentIdentifier := db.CommentIdentifier{
// 				Identifier: alreadyExistingCommentIdentifier.Identifier,
// 				UserID:     token.UID,
// 				PostID:     req.PostID,
// 			}
// 			fmt.Println("POINT 66")
// 			if req.ParentCommentID != nil {
// 				resaveCommentIdentifier.ParentIdentifier = futureParentIdentifier.Identifier
// 			}
// 			fmt.Println("POINT 7")
// 			err = tx.
// 				Table("comment_identifiers").
// 				Where("user_id = ?", token.UID).
// 				Where("identifier = ?", alreadyExistingCommentIdentifier.Identifier).
// 				Where("post_id = ?", req.PostID).
// 				Where("parent_identifier = ?", resaveCommentIdentifier.ParentIdentifier).
// 				Error
// 			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
// 				tx.Rollback()
// 				response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 				return
// 			} else if errors.Is(err, gorm.ErrRecordNotFound) {
// 				// todo: insert and do this with new id?
// 				err = tx.Create(&resaveCommentIdentifier).Error
// 				if err != nil {
// 					tx.Rollback()
// 					response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 					return
// 				}
// 				comment.IdentifierID = resaveCommentIdentifier.ID
// 			} else {
// 				comment.IdentifierID = *alreadyExistingCommentIdentifier.Identifier

// 			}
// 			fmt.Println("POINT 10")
// 			// note: the id of this record, and if there is a keyerror based on unique (user_id, post_id, identifier, parent_identifer) then we just set the comment pointer to that id
// 		}

// 	}

// 	// save the comment
// 	err = tx.Create(&comment).Error
// 	if err != nil {
// 		fmt.Println("ROLL BACK")

// 		tx.Rollback()
// 		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 		return
// 	}

// 	// if all goes well, respond with a 201 & commit the transaction
// 	err = tx.Commit().Error
// 	if err != nil {
// 		fmt.Println("ROLL BACK")

// 		tx.Rollback()
// 		response.New(http.StatusInternalServerError).Err(serverError.Error()).Send(c)
// 		return
// 	}
// 	response.New(http.StatusCreated).Send(c)
// }
