set tabstop=4 shiftwidth=4 expandtab
" set tabstop=2 shiftwidth=2 expandtab " for arduino C

set autoindent
set hlsearch
set maxmempattern=2000

filetype plugin indent on
autocmd FileType yaml setlocal ts=2 sts=2 sw=2 expandtab

if has("autocmd")
  au BufReadPost * if line("'\"") > 0 && line("'\"") <= line("$") | exe "normal! g`\"" | endif
endif
