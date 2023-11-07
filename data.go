package main

var routes = []Site{
	{
		Handlers: defaultHandlers,
		Name:     "about",
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
		Handlers:  defaultHandlers,
		ServePath: "/Users/rxlx/ui/",
		Name:      "contact",
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
		Handlers: defaultHandlers,
		Name:     "blog",
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
		Handlers: defaultHandlers,
		Name:     "news",
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
