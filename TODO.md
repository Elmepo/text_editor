# Todo:

- [ ] Create a parent class for text contents, and then pseudo-inherit this for command vs file contents?
- [ ] Probably need to come up with a better way to run commands
- [x] Fix backspace
- [x] Fix arrow keys
- [x] Fix backspace across lines
- [x] backspace cursor positioning is pretty bad?
- [x] fix delete and backspace in command buffers
- [x] Look into preventing the automatic overflow for characters to handle it within the text editor (i.e. overflow y, line break, etc)
    - [x] Prevent automatic overflow
    - [?] Shift "window" to always display the cursor
    - [x] setting to automatically wrap lines
    - [ ] configurable "window"
- [x] Automatically resize text across lines
- [x] save file should be able to take a parameter for the file name
- [x] should be able to automatically open up a file for writing and this is the save default
- [ ] log lines for commands that are pseudo permanent, kinda like vim
- [ ] better indicators for command mode vs insert mode
- [x] if running with an existing file name as the param, open for reading
- [x] Can open file into buffer for reading
- [x] delineate between command and entry zones via a unicode half width block or similar (u+2584)
- [x] show page via unicode right half block (u+2590)
- [ ] wrapping should happen on word not on character - i.e. it should be "a wrapped\nstring" instead of "a wrapped st\nring"
- [ ] cannot read the previous line when using left on 0 position. Might need to store a relative number of lines to the current cursor instead?
- [ ] In general I think how the relationship between stored text and the displayed/rendered text is wrong and needs to be massively improved. too much hacking at the edges right now
- [ ] Rendering Improvements:
    - [ ] Only render what's important: Render Line numbers and debug information only at the start and (eventually) when the page grows to be more than the terminal size and the cursor moves down enough.
    - [ ] Only render text updates, i.e. don't re-render on every cursor movement?
- [ ] Think I should change from `switch char: if command_mode ? command logic : file logic` to `if command_mode ? switch char : switch char`? Really just thinking I might need to have the logic in two seperate files/receviers entirely.
