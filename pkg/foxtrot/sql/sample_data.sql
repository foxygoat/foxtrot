INSERT INTO rooms (name) VALUES
	('$Kitchen'),
	('$Shed');

INSERT INTO users (name, password_hash) VALUES
	('$Fox', '$2a$10$V5.UzTYmeYh.bPz51WiIH.Yp2KawEqEmgF/amTTXtOHBvcjkFuIrC'),   -- Password: Pa$$w0rd
	('$Goat', '$2a$10$Rp2rFH12j0Ovc8VUfJKEX.O2SKHDpHs1b6KBkCqluzSMOowuDagk2'),  -- Password: $s3cr37
	('$Cat', '$2a$10$V5.UzTYmeYh.bPz51WiIH.Yp2KawEqEmgF/amTTXtOHBvcjkFuIrC'),   -- Password: Pa$$w0rd
	('$Camel', '$2a$10$Rp2rFH12j0Ovc8VUfJKEX.O2SKHDpHs1b6KBkCqluzSMOowuDagk2'); -- Password: $s3cr37

INSERT INTO messages (id, content, created_at, room, author) VALUES
	(1, 'Hi', '2020-11-22T11:11:11Z', '$Kitchen', '$Goat'),
	(2, 'Hallo', '2020-11-22T11:12:12Z', '$Kitchen', '$Fox'),
	(3, 'Are you hungry?', '2020-11-22T12:22:42Z', '$Kitchen', '$Goat'),
	(4, 'Yes.', '2020-11-22T12:32:42Z', '$Kitchen', '$Fox'),
	(5, 'Ok, bye.', '2020-11-22T13:00:42Z', '$Kitchen', '$Goat'),
	(6, 'Hi', '2020-11-22T13:11:11Z', '$Shed', '$Goat'),
	(7, 'Hallo', '2020-11-22T13:12:12Z', '$Shed', '$Cat'),
	(8, 'Howdie', '2020-11-22T13:22:42Z', '$Shed', '$Camel'),
	(9, "Don't go into the kitchen - there's a hungry fox.", '2020-11-22T13:32:42Z', '$Shed', '$Goat'),
	(10, 'Ok, thanks.', '2020-11-22T13:33:42Z', '$Shed', '$Cat'),
	(11, 'I eat fox for breakfast.', '2020-11-22T13:34:42Z', '$Shed', '$Camel'),
	(100, 'Sentinel message with ID 100 - eat that.', '2020-11-22T23:44:44Z', '$Shed', '$Camel');
