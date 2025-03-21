-- This is the entry project local nvim configuration
-- used by plugin "klen/nvim-config-local" to
-- load project specific settings, when nvim is
-- started on this project.

-- Load debug adapters in this project.
-- Needs https://github.com/ldelossa/nvim-dap-projects
nvimdap = require("nvim-dap-projects")
nvimdap.search_project_config()
