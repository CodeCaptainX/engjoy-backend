package sentencepack

type SeedSentence struct {
	Text     string
	Source   string
	Category string
}

var EnvironmentSentencePack = []SeedSentence{
	{Text: "Please open your textbook to page twelve.", Source: "starter-pack", Category: "school"},
	{Text: "Can you explain that question one more time?", Source: "starter-pack", Category: "school"},
	{Text: "I forgot my homework at home this morning.", Source: "starter-pack", Category: "school"},
	{Text: "We have a science test after lunch.", Source: "starter-pack", Category: "school"},
	{Text: "Could I borrow a pen for this worksheet?", Source: "starter-pack", Category: "school"},
	{Text: "The nurse will call your name in a few minutes.", Source: "starter-pack", Category: "hospital"},
	{Text: "I have had a sore throat since yesterday.", Source: "starter-pack", Category: "hospital"},
	{Text: "Do you have any allergies or medical conditions?", Source: "starter-pack", Category: "hospital"},
	{Text: "Please take this medicine after meals.", Source: "starter-pack", Category: "hospital"},
	{Text: "You should get some rest and drink plenty of water.", Source: "starter-pack", Category: "hospital"},
	{Text: "Pass the ball and move into space.", Source: "starter-pack", Category: "sports"},
	{Text: "Our team needs to defend better in the second half.", Source: "starter-pack", Category: "sports"},
	{Text: "The coach wants us to warm up before practice.", Source: "starter-pack", Category: "sports"},
	{Text: "She scored the winning goal at the end of the match.", Source: "starter-pack", Category: "sports"},
	{Text: "Keep your eyes on the ball and follow through.", Source: "starter-pack", Category: "sports"},
	{Text: "Could we see the menu, please?", Source: "starter-pack", Category: "restaurant"},
	{Text: "I would like a bowl of soup and a glass of water.", Source: "starter-pack", Category: "restaurant"},
	{Text: "Is this dish spicy or mild?", Source: "starter-pack", Category: "restaurant"},
	{Text: "Can we have the bill when you have a moment?", Source: "starter-pack", Category: "restaurant"},
	{Text: "Do you have any vegetarian options today?", Source: "starter-pack", Category: "restaurant"},
	{Text: "What time does the next train leave?", Source: "starter-pack", Category: "travel"},
	{Text: "I would like a ticket to the city center.", Source: "starter-pack", Category: "travel"},
	{Text: "Could you show me where platform three is?", Source: "starter-pack", Category: "travel"},
	{Text: "We should leave early to avoid traffic.", Source: "starter-pack", Category: "travel"},
	{Text: "This bus stops near the museum, right?", Source: "starter-pack", Category: "travel"},
	{Text: "How much does this cost altogether?", Source: "starter-pack", Category: "shopping"},
	{Text: "Do you have this in a different size?", Source: "starter-pack", Category: "shopping"},
	{Text: "I am just looking, thank you.", Source: "starter-pack", Category: "shopping"},
	{Text: "Can I pay by card here?", Source: "starter-pack", Category: "shopping"},
	{Text: "The cashier gave me the wrong change.", Source: "starter-pack", Category: "shopping"},
	{Text: "Could you send me the updated schedule by email?", Source: "starter-pack", Category: "office"},
	{Text: "Let us review the report before the meeting starts.", Source: "starter-pack", Category: "office"},
	{Text: "I need a little more time to finish this task.", Source: "starter-pack", Category: "office"},
	{Text: "The manager asked for a quick progress update.", Source: "starter-pack", Category: "office"},
	{Text: "Can we move this discussion to tomorrow morning?", Source: "starter-pack", Category: "office"},
	{Text: "Excuse me, could you help me find this address?", Source: "starter-pack", Category: "daily-life"},
	{Text: "I think I left my phone on the table.", Source: "starter-pack", Category: "daily-life"},
	{Text: "It is starting to rain, so bring an umbrella.", Source: "starter-pack", Category: "daily-life"},
	{Text: "Please wait here for a moment.", Source: "starter-pack", Category: "daily-life"},
	{Text: "What do people usually say in this situation?", Source: "starter-pack", Category: "daily-life"},
}
