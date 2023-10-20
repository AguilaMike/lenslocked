package controllers

import (
	b64 "encoding/base64"
	"errors"
	"fmt"
	"math/rand"
	"net/http"

	"github.com/AguilaMike/lenslocked/pkg/app/context"
	"github.com/AguilaMike/lenslocked/pkg/app/models"
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
	ID     uuid.UUID
	UserID uuid.UUID
	Title  string
	ID64   string
	Images []string
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

	gallery, err := g.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		g.Templates.New.Execute(w, r, data, err)
		return
	}
	data.ID = gallery.ID
	editPath := fmt.Sprintf("/galleries/%s/edit", data.IDEncode())
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
	g.Templates.Edit.Execute(w, r, data)
}

func (g Galleries) Update(w http.ResponseWriter, r *http.Request) {
	var data GalleryDTO
	_, err := data.IDDecodeFromString(chi.URLParam(r, "id"))
	gallery, ok := g.validate(w, r, &data, err, userMustOwnGallery)
	if !ok {
		return
	}
	data.Title = r.FormValue("title")
	gallery.Title = data.Title
	err = g.GalleryService.Update(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}
	editPath := fmt.Sprintf("/galleries/%s/edit", data.IDEncode())
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
	gallery, ok := g.validate(w, r, &data, err)
	if !ok {
		return
	}
	data.ID = gallery.ID
	data.Title = gallery.Title

	// We are going to psuedo-randomly come up with 20 images to render for our
	// gallery until we actually support uploading images. These images will use
	// placekitten.com, which gives us cat images.
	for i := 0; i < 10; i++ {
		// width and height are random values betwee 200 and 700
		w, h := rand.Intn(500)+200, rand.Intn(500)+200
		// using the width and height, we generate a URL
		catImageURL := fmt.Sprintf("https://placekitten.com/%d/%d", w, h)
		// Then we add the URL to our images.
		data.Images = append(data.Images, catImageURL)
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
