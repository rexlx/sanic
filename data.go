package main

var routes = []Site{
	{
		Name: "about",
		UI: UIConfig{
			Style: BasicStyle{
				BodyBG:   "#f5f5f5",
				BodyText: "#333",
				H1:       "#444",
				Btn:      "#becdc3",
				BtnText:  "#000",
			},
			Templates: []Template{
				{
					Name: "index",
					Body: splashPage,
				},
			},
		},
	},
	{
		Name: "contact",
		UI: UIConfig{
			Style: BasicStyle{
				BodyBG:   "#f5f5f5",
				BodyText: "#333",
				H1:       "#444",
				Btn:      "#708e93",
				BtnText:  "#000",
			},
			Templates: []Template{
				{
					Name: "index",
					Body: splashPage,
				},
			},
		},
	},
	{
		Name: "blog",
		UI: UIConfig{
			Style: BasicStyle{
				BodyBG:   "#828599",
				BodyText: "#333",
				H1:       "#444",
				Btn:      "#becdc3",
				BtnText:  "#000",
			},
			Templates: []Template{
				{
					Name: "index",
					Body: splashPage,
				},
			},
		},
	},
	{
		Name: "news",
		UI: UIConfig{
			Style: BasicStyle{
				BodyBG:   "#333",
				BodyText: "#828599",
				H1:       "#444",
				Btn:      "#becdc3",
				BtnText:  "#000",
			},
			Templates: []Template{
				{
					Name: "index",
					Body: splashPage,
				},
			},
		},
	},
}
