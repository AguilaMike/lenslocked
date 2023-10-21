package controllers

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"

	"github.com/AguilaMike/lenslocked/pkg/app/context"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
	"github.com/AguilaMike/lenslocked/pkg/internal/utils"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Galleries struct {
	Templates struct {
		Show  Template
		New   Template
		Edit  Template
		Index Template
	}
	GalleryService *models.GalleryService
}

type GalleryDTO struct {
	ID     uuid.UUID `form:"id"`
	UserID uuid.UUID
	Title  string  `form:"title"`
	Public bool    `form:"public"`
	ID64   string  `form:"id64"`
	Images []Image `form:"images"`
}

type Image struct {
	GalleryID       string
	Filename        string
	FilenameEscaped string
}

func (g *GalleryDTO) IDEncode() string {
	g.ID64 = b64.StdEncoding.EncodeToString([]byte(g.ID.String()))
	return g.ID64
}

func (g *GalleryDTO) IDDecodeFromBase64String(id string) (uuid.UUID, error) {
	decodeID, _ := b64.StdEncoding.DecodeString(id)
	return g.IDDecodeFromString(string(decodeID))
}

func (g *GalleryDTO) IDDecodeFromString(id string) (uuid.UUID, error) {
	var err error
	g.ID, err = uuid.Parse(id)
	return g.ID, err
}

func (g Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	data.Title = r.FormValue("title")
	g.Templates.New.Execute(w, r, data)
}

func (g Galleries) Create(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	data.UserID = context.User(r.Context()).ID
	data.Title = r.FormValue("title")
	data.Public = utils.ConvertBoolCheckbox(r.FormValue("public"))

	gallery, err := g.GalleryService.Create(data.Title, data.UserID, data.Public)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}
	data.ID = gallery.ID
	// editPath := fmt.Sprintf("/galleries/%s/edit", data.IDEncode())
	editPath := fmt.Sprintf("/galleries")
	http.Redirect(w, r, editPath, http.StatusFound)
}

func userMustOwnGallery(w http.ResponseWriter, r *http.Request, data *GalleryDTO, gallery *models.Gallery) error {
	data.UserID = context.User(r.Context()).ID
	if data.UserID != gallery.UserID {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}
	return nil
}

func userMustPrivateGallery(w http.ResponseWriter, r *http.Request, data *GalleryDTO, gallery *models.Gallery) error {
	user := context.User(r.Context())
	if user != nil && user.ID != uuid.Nil {
		data.UserID = context.User(r.Context()).ID
	} else {
		data.UserID = uuid.Nil
	}

	if data.UserID != gallery.UserID && !gallery.Public {
		http.Error(w, "You are not authorized to see this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}
	return nil
}

type galleryOpt func(http.ResponseWriter, *http.Request, *GalleryDTO, *models.Gallery) error

func (g Galleries) validate(w http.ResponseWriter, r *http.Request, data *GalleryDTO, err error, opts ...galleryOpt) (*models.Gallery, bool) {
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return nil, false
	}
	gallery, err := g.GalleryService.ByID(data.ID)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return nil, false
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, false
	}
	for _, opt := range opts {
		err = opt(w, r, data, gallery)
		if err != nil {
			return nil, false
		}
	}
	return gallery, true
}

func (g Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	_, err := data.IDDecodeFromBase64String(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustOwnGallery)
	if !ok {
		return
	}
	data.Title = gallery.Title
	data.Public = gallery.Public
	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}
	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	_, err := data.IDDecodeFromString(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustOwnGallery)
	if !ok {
		return
	}
	r.ParseForm()
	data.Title = r.FormValue("title")
	data.Public = utils.ConvertBoolCheckbox(r.FormValue("public"))
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	gallery.Title = data.Title
	gallery.Public = data.Public
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	// editPath := fmt.Sprintf("/galleries/%s/edit", data.IDEncode())
	editPath := fmt.Sprintf("/galleries")
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) Index(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Galleries []GalleryDTO
	}

	user := context.User(r.Context()).ID
	galleries, err := g.GalleryService.ByUserID(user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	for _, gallery := range galleries {
		item := GalleryDTO{
			ID:     gallery.ID,
			UserID: user,
			Title:  gallery.Title,
			Public: gallery.Public,
		}
		item.IDEncode()
		data.Galleries = append(data.Galleries, item)
	}
	// TODO: Lookup the galleries we are going to render
	g.Templates.Index.Execute(w, r, data)
}

func (g Galleries) Show(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	_, err := data.IDDecodeFromBase64String(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustPrivateGallery)
	if !ok {
		return
	}
	data.ID = gallery.ID
	data.Title = gallery.Title

	images, err := g.GalleryService.Images(gallery.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	for _, image := range images {
		data.Images = append(data.Images, Image{
			GalleryID:       image.GalleryID,
			Filename:        image.Filename,
			FilenameEscaped: url.PathEscape(image.Filename),
		})
	}

	g.Templates.Show.Execute(w, r, data)
}

func (g Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	_, err := data.IDDecodeFromBase64String(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustOwnGallery)
	if !ok {
		return
	}
	err = g.GalleryService.Delete(gallery.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/galleries", http.StatusFound)
}

func (g Galleries) filename(w http.ResponseWriter, r *http.Request) string {
	filename := chi.URLParam(r, "filename")
	filename = filepath.Base(filename)
	return filename
}

func (g Galleries) Image(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	var data GalleryDTO
	_, err := data.IDDecodeFromString(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return
	}
	// images, err := g.GalleryService.Images(data.ID)
	// if err != nil {
	// 	fmt.Println(err)
	// 	http.Error(w, "Something went wrong", http.StatusInternalServerError)
	// 	return
	// }
	// var requestedImage models.Image
	// imageFound := false
	// for _, image := range images {
	// 	if image.Filename == filename {
	// 		requestedImage = image
	// 		imageFound = true
	// 		break
	// 	}
	// }
	// if !imageFound {
	// 	http.Error(w, "Image not found", http.StatusNotFound)
	// 	return
	// }
	image, err := g.GalleryService.Image(data.ID, filename)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Image not found", http.StatusNotFound)
			return
		}
		fmt.Println(err)
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	http.ServeFile(w, r, image.Path)
}

func (g Galleries) DeleteImage(w http.ResponseWriter, r *http.Request) {
	filename := g.filename(w, r)
	var data GalleryDTO
	_, err := data.IDDecodeFromString(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustOwnGallery)
	if !ok {
		return
	}
	err = g.GalleryService.DeleteImage(gallery.ID, filename)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%s/edit", data.IDEncode())
	http.Redirect(w, r, editPath, http.StatusFound)
}

func (g Galleries) UploadImage(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	_, err := data.IDDecodeFromString(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustOwnGallery)
	if !ok {
		return
	}
	err = r.ParseMultipartForm(5 << 20) // 5mb
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	fileHeaders := r.MultipartForm.File["images"]
	for _, fileHeader := range fileHeaders {
		file, err := fileHeader.Open()
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
		defer file.Close()
		//fmt.Printf("Attempting to upload %v for gallery %s.\n", fileHeader.Filename, gallery.ID.String())
		// io.Copy(w, file)
		err = g.GalleryService.CreateImage(gallery.ID, fileHeader.Filename, file)
		if err != nil {
			http.Error(w, "Something went wrong", http.StatusInternalServerError)
			return
		}
	}
	editPath := fmt.Sprintf("/galleries/%s/edit", data.IDEncode())
	http.Redirect(w, r, editPath, http.StatusFound)
}
