package crud

import "context"

func (r *Repo) GetNameForId(ctx context.Context, id string) (string, error) {
	name, found := r.NameCache.Get(id)
	if !found {
		if player, err := r.PlayerCrud.GetById(ctx, id, nil); err == nil {
			r.NameCache.Set(id, player.Name)
			return player.Name, nil
		} else {
			if place, err := r.PlaceCrud.GetById(ctx, id, nil); err == nil {
				r.NameCache.Set(id, place.Name)
				return place.Name, nil
			} else {
				if coach, err := r.CoachCrud.GetById(ctx, id, nil); err == nil {
					r.NameCache.Set(id, coach.Name)
					return coach.Name, nil
				} else {
					return "", err
				}
			}
		}
	} else {
		return name, nil
	}
}
