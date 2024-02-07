package controllers

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/taherk/galleryapp/context"
	"github.com/taherk/galleryapp/models"
)

type Galleries struct {
	Templates struct {
		New   Template
		Edit  Template
		Index Template
		Show  Template
	}

	GalleryService *models.GalleryService
}

func (ctrl Galleries) New(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title string
	}
	data.Title = r.FormValue("title")
	ctrl.Templates.New.Execute(w, r, data)
}

func (ctrl Galleries) Create(w http.ResponseWriter, r *http.Request) {
	// create gallery
	var data struct {
		Title  string
		UserID int
	}
	data.UserID = int(context.User(r.Context()).ID)
	data.Title = r.FormValue("title")

	gallery, err := ctrl.GalleryService.Create(data.Title, data.UserID)
	if err != nil {
		// if failed redirect back to the create page
		ctrl.Templates.New.Execute(w, r, data, err)
		return
	}

	// if created redirect to edit page
	http.Redirect(w, r, fmt.Sprintf("/galleries/%d/edit", gallery.ID), http.StatusFound)
}

func (ctrl Galleries) Edit(w http.ResponseWriter, r *http.Request) {
	gallery, err := ctrl.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	// render the edit page
	data := struct {
		ID    int
		Title string
	}{
		ID:    gallery.ID,
		Title: gallery.Title,
	}
	ctrl.Templates.Edit.Execute(w, r, data)
}

func (ctrl Galleries) Show(w http.ResponseWriter, r *http.Request) {
	gallery, err := ctrl.galleryByID(w, r)
	if err != nil {
		return
	}

	var data struct {
		ID     int
		Title  string
		Images []string
	}
	data.ID = gallery.ID
	data.Title = gallery.Title
	for i := 0; i < 20; i++ {
		w, h := rand.Intn(500)+200, rand.Intn(500)+200
		catImageURL := fmt.Sprintf("https://placekitten.com/%d/%d", w, h)
		data.Images = append(data.Images, catImageURL)
	}

	ctrl.Templates.Show.Execute(w, r, data)
}

func (ctrl Galleries) Update(w http.ResponseWriter, r *http.Request) {
	gallery, err := ctrl.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	gallery.Title = r.FormValue("title")
	err = ctrl.GalleryService.UpdateTitle(gallery)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, fmt.Sprintf("/galleries/%d/edit", gallery.ID), http.StatusFound)
}

func (ctrl Galleries) Index(w http.ResponseWriter, r *http.Request) {
	type Gallery struct {
		ID    int
		Title string
	}

	var data struct {
		Galleries []Gallery
	}

	user := context.User(r.Context())
	galleries, err := ctrl.GalleryService.ByUserID(int(user.ID))
	if err != nil {
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	for _, gallery := range galleries {
		data.Galleries = append(data.Galleries, Gallery{
			ID:    gallery.ID,
			Title: gallery.Title,
		})
	}

	ctrl.Templates.Index.Execute(w, r, data)
}

func (ctrl Galleries) Delete(w http.ResponseWriter, r *http.Request) {
	gallery, err := ctrl.galleryByID(w, r, userMustOwnGallery)
	if err != nil {
		return
	}

	// render the edit page
	err = ctrl.GalleryService.Delete(gallery.ID)
	if err != nil {
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/galleries", http.StatusFound)
}

type galleryOpt func(http.ResponseWriter, *http.Request, *models.Gallery) error

func (ctrl Galleries) galleryByID(w http.ResponseWriter, r *http.Request, options ...galleryOpt) (*models.Gallery, error) {

	// validate gallery id to be int
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusNotFound)
		return nil, err
	}

	// get gallery
	gallery, err := ctrl.GalleryService.ByID(id)
	if err != nil {
		if errors.Is(err, models.ErrNotFound) {
			http.Error(w, "Gallery not found", http.StatusNotFound)
			return nil, err
		}
		http.Error(w, "Something went wrong", http.StatusInternalServerError)
		return nil, err
	}

	for _, opt := range options {
		err = opt(w, r, gallery)
		if err != nil {
			return nil, err
		}
	}

	return gallery, nil
}

func userMustOwnGallery(w http.ResponseWriter, r *http.Request, gallery *models.Gallery) error {
	user := context.User(r.Context())
	if gallery.UserID != int(user.ID) {
		http.Error(w, "You are not authorized to edit this gallery", http.StatusForbidden)
		return fmt.Errorf("user does not have access to this gallery")
	}

	return nil
}
