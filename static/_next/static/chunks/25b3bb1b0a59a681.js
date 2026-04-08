(globalThis.TURBOPACK||(globalThis.TURBOPACK=[])).push(["object"==typeof document?document.currentScript:void 0,36791,e=>{"use strict";var r=e.i(47167),t=e.i(43476),s=e.i(71645),a=e.i(85665),o=e.i(64556),n=e.i(59161);let l=s.forwardRef(function({title:e,titleId:r,...t},a){return s.createElement("svg",Object.assign({xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 24 24",strokeWidth:1.5,stroke:"currentColor","aria-hidden":"true","data-slot":"icon",ref:a,"aria-labelledby":r},t),e?s.createElement("title",{id:r},e):null,s.createElement("path",{strokeLinecap:"round",strokeLinejoin:"round",d:"m8.25 4.5 7.5 7.5-7.5 7.5"}))});var i=e.i(70744),d=e.i(87906),c=e.i(39875),x=e.i(91929),m=e.i(24189);let h=s.forwardRef(function({title:e,titleId:r,...t},a){return s.createElement("svg",Object.assign({xmlns:"http://www.w3.org/2000/svg",fill:"none",viewBox:"0 0 24 24",strokeWidth:1.5,stroke:"currentColor","aria-hidden":"true","data-slot":"icon",ref:a,"aria-labelledby":r},t),e?s.createElement("title",{id:r},e):null,s.createElement("path",{strokeLinecap:"round",strokeLinejoin:"round",d:"m3.75 13.5 10.5-11.25L12 10.5h8.25L9.75 21.75 12 13.5H3.75Z"}))});var p=e.i(15877),u=e.i(7657),g=e.i(84174);let b=[{category:"URL Shortening",method:"POST",path:"/api/v1/shorten",description:"Shorten a URL",auth:"Optional",requestBody:{contentType:"application/json",schema:`{
  "url": "string (required)",
  "custom_code": "string (optional)",
  "expires_at": "string (optional, ISO 8601 format)"
}`},responses:[{status:201,description:"URL shortened successfully",schema:`{
  "short_code": "string",
  "short_url": "string",
  "original_url": "string",
  "created_at": "string",
  "expires_at": "string | null"
}`},{status:400,description:"Invalid request"},{status:409,description:"Custom code already exists"}],example:{request:`{
  "url": "https://example.com/very-long-url",
  "custom_code": "my-link"
}`,response:`{
  "short_code": "my-link",
  "short_url": "${r.default.env.NEXT_PUBLIC_API_URL||"http://localhost:8080"}/my-link",
  "original_url": "https://example.com/very-long-url",
  "created_at": "2023-01-01T12:00:00Z"
}`}},{category:"URL Shortening",method:"GET",path:"/{shortCode}",description:"Redirect to original URL",parameters:[{name:"shortCode",type:"string",required:!0,description:"The short code for the URL"}],responses:[{status:301,description:"Redirects to original URL"},{status:404,description:"URL not found"},{status:410,description:"URL expired"}]},{category:"URL Shortening",method:"POST",path:"/api/v1/batch-shorten",description:"Shorten multiple URLs at once",auth:"Required",requestBody:{contentType:"application/json",schema:`[
  {
    "url": "string (required)",
    "custom_code": "string (optional)"
  }
]`},responses:[{status:200,description:"Batch processing completed",schema:`{
  "successful": "number",
  "failed": "number",
  "results": "Array of shortened URLs",
  "errors": "Array of errors"
}`},{status:401,description:"Authentication required"}]},{category:"URL Shortening",method:"GET",path:"/api/v1/my-urls",description:"Get user's URLs",auth:"Required",parameters:[{name:"page",type:"number",required:!1,description:"Page number (default: 1)"},{name:"limit",type:"number",required:!1,description:"Items per page (default: 20, max: 100)"}],responses:[{status:200,description:"User URLs retrieved successfully",schema:`{
  "urls": "Array of URL objects",
  "total": "number",
  "page": "number",
  "limit": "number",
  "total_pages": "number"
}`},{status:401,description:"Authentication required"}]},{category:"URL Shortening",method:"PUT",path:"/api/v1/my-urls/{shortCode}",description:"Update a shortened URL",auth:"Required",parameters:[{name:"shortCode",type:"string",required:!0,description:"The short code to update"}],requestBody:{contentType:"application/json",schema:`{
  "original_url": "string (optional)",
  "expires_at": "string (optional)"
}`},responses:[{status:200,description:"URL updated successfully"},{status:404,description:"URL not found"}]},{category:"URL Shortening",method:"DELETE",path:"/api/v1/my-urls/{shortCode}",description:"Delete a shortened URL",auth:"Required",parameters:[{name:"shortCode",type:"string",required:!0,description:"The short code to delete"}],responses:[{status:200,description:"URL deleted successfully"},{status:404,description:"URL not found"}]},{category:"System",method:"GET",path:"/health",description:"Health check endpoint",responses:[{status:200,description:"System is healthy",schema:`{
  "status": "healthy",
  "timestamp": "string"
}`}]}];function y(){let[e,y]=(0,s.useState)(null),[f,j]=(0,s.useState)(null),[N,v]=(0,s.useState)("All"),[w,k]=(0,s.useState)(""),[U,R]=(0,s.useState)("http://localhost:8080"),[L,C]=(0,s.useState)("javascript");(0,s.useEffect)(()=>{R(window.location.origin.includes("localhost")?"http://localhost:8080":r.default.env.NEXT_PUBLIC_API_URL||"https://api.trunc.site")},[]);let _=(0,s.useMemo)(()=>["All",...new Set(b.map(e=>e.category||"Other"))],[]),I=(0,s.useMemo)(()=>{let e=b;return"All"!==N&&(e=e.filter(e=>e.category===N)),w&&(e=e.filter(e=>e.path.toLowerCase().includes(w.toLowerCase())||e.description.toLowerCase().includes(w.toLowerCase())||e.method.toLowerCase().includes(w.toLowerCase()))),e},[N,w]),S=(e,r)=>{navigator.clipboard.writeText(e).then(()=>{j(r),setTimeout(()=>j(null),2e3)})},A=e=>{switch(e){case"GET":return"bg-green-500 text-white";case"POST":return"bg-blue-500 text-white";case"PUT":return"bg-yellow-500 text-white";case"DELETE":return"bg-red-500 text-white";default:return"bg-gray-500 text-white"}},E=e=>{switch(e){case"URL Shortening":return(0,t.jsx)(h,{className:"h-4 w-4"});case"API Keys":return(0,t.jsx)(m.KeyIcon,{className:"h-4 w-4"});case"System":return(0,t.jsx)(p.ServerIcon,{className:"h-4 w-4"});default:return(0,t.jsx)(n.DocumentTextIcon,{className:"h-4 w-4"})}};return(0,t.jsx)(a.default,{children:(0,t.jsxs)("div",{className:"min-h-screen bg-gradient-to-br from-background via-background to-muted/20",children:[(0,t.jsx)("div",{className:"bg-gradient-to-br from-primary via-primary/90 to-primary/80 text-primary-foreground",children:(0,t.jsxs)("div",{className:"max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12",children:[(0,t.jsxs)("div",{className:"flex items-center gap-4 mb-6",children:[(0,t.jsx)("div",{className:"p-3 bg-background/10 backdrop-blur-sm rounded-xl",children:(0,t.jsx)(n.DocumentTextIcon,{className:"h-10 w-10"})}),(0,t.jsxs)("div",{children:[(0,t.jsx)("h1",{className:"text-4xl font-bold mb-2",children:"API Documentation"}),(0,t.jsx)("p",{className:"text-primary-foreground/80 text-lg",children:"Complete reference for integrating Trunc into your applications"})]})]}),(0,t.jsxs)("div",{className:"grid md:grid-cols-3 gap-4 mt-8",children:[(0,t.jsxs)("div",{className:"bg-background/10 backdrop-blur-sm rounded-xl p-6 border border-background/20 hover:bg-background/15 transition-all duration-300",children:[(0,t.jsxs)("div",{className:"flex items-center gap-3 mb-3",children:[(0,t.jsx)(x.GlobeAltIcon,{className:"h-6 w-6 text-primary-foreground/80"}),(0,t.jsx)("h3",{className:"font-semibold text-lg",children:"Base URL"})]}),(0,t.jsx)("code",{className:"block bg-black/20 backdrop-blur-sm px-4 py-3 rounded-lg text-sm font-mono text-primary-foreground",children:U}),(0,t.jsxs)("p",{className:"text-xs text-primary-foreground/80 mt-3",children:["Production: ",(0,t.jsx)("code",{className:"bg-black/20 px-2 py-1 rounded",children:"https://api.trunc.site"})]})]}),(0,t.jsxs)("div",{className:"bg-background/10 backdrop-blur-sm rounded-xl p-6 border border-background/20 hover:bg-background/15 transition-all duration-300",children:[(0,t.jsxs)("div",{className:"flex items-center gap-3 mb-3",children:[(0,t.jsx)(u.LockClosedIcon,{className:"h-6 w-6 text-primary-foreground/80"}),(0,t.jsx)("h3",{className:"font-semibold text-lg",children:"Authentication"})]}),(0,t.jsx)("code",{className:"block bg-black/20 backdrop-blur-sm px-4 py-3 rounded-lg text-sm font-mono text-primary-foreground break-all",children:"Authorization: Bearer YOUR_KEY"}),(0,t.jsxs)("p",{className:"text-xs text-primary-foreground/80 mt-3",children:["API keys start with ",(0,t.jsx)("code",{className:"bg-black/20 px-2 py-1 rounded",children:"us_"})," prefix"]})]}),(0,t.jsxs)("div",{className:"bg-background/10 backdrop-blur-sm rounded-xl p-6 border border-background/20 hover:bg-background/15 transition-all duration-300",children:[(0,t.jsxs)("div",{className:"flex items-center gap-3 mb-3",children:[(0,t.jsx)(g.SparklesIcon,{className:"h-6 w-6 text-primary-foreground/80"}),(0,t.jsx)("h3",{className:"font-semibold text-lg",children:"Rate Limits"})]}),(0,t.jsxs)("div",{className:"space-y-2 text-sm",children:[(0,t.jsxs)("div",{className:"flex justify-between items-center bg-black/20 backdrop-blur-sm px-4 py-2 rounded-lg",children:[(0,t.jsx)("span",{children:"Standard"}),(0,t.jsx)("code",{className:"font-mono",children:"1000/hour"})]}),(0,t.jsx)("p",{className:"text-xs text-primary-foreground/80 mt-2",children:"Configurable per API key"})]})]})]})]})}),(0,t.jsxs)("div",{className:"max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8",children:[(0,t.jsxs)("div",{className:"mb-8 space-y-4",children:[(0,t.jsxs)("div",{className:"relative",children:[(0,t.jsx)(c.MagnifyingGlassIcon,{className:"absolute left-4 top-1/2 transform -translate-y-1/2 h-5 w-5 text-muted-foreground"}),(0,t.jsx)("input",{type:"text",placeholder:"Search endpoints by name, path, or method...",value:w,onChange:e=>k(e.target.value),className:"w-full pl-12 pr-4 py-4 rounded-xl border-2 border-border bg-card text-foreground focus:border-primary focus:ring-4 focus:ring-primary/20 transition-all outline-none placeholder-muted-foreground"})]}),(0,t.jsx)("div",{className:"flex flex-wrap gap-3",children:_.map(e=>(0,t.jsxs)("button",{onClick:()=>v(e),className:`group px-5 py-2.5 rounded-xl text-sm font-medium transition-all duration-300 flex items-center gap-2 ${N===e?"bg-primary text-primary-foreground shadow-lg shadow-primary/20 scale-105":"bg-card text-foreground hover:bg-muted border-2 border-border hover:border-primary/50"}`,children:["All"!==e&&E(e),e,(0,t.jsx)("span",{className:`px-2 py-0.5 rounded-full text-xs font-semibold ${N===e?"bg-background/20 text-primary-foreground":"bg-muted text-muted-foreground group-hover:bg-primary/10 group-hover:text-primary"}`,children:b.filter(r=>"All"===e||r.category===e).length})]},e))})]}),(0,t.jsxs)("div",{className:"grid grid-cols-1 lg:grid-cols-3 gap-8",children:[(0,t.jsx)("div",{className:"lg:col-span-1",children:(0,t.jsxs)("div",{className:"bg-card rounded-2xl shadow-xl border border-border p-6 sticky top-8 max-h-[calc(100vh-8rem)] overflow-y-auto",children:[(0,t.jsxs)("div",{className:"flex items-center justify-between mb-6",children:[(0,t.jsxs)("h2",{className:"text-xl font-bold text-foreground flex items-center gap-2",children:[(0,t.jsx)(o.CodeBracketIcon,{className:"h-6 w-6 text-primary"}),"Endpoints"]}),(0,t.jsx)("span",{className:"px-3 py-1 bg-primary/10 text-primary rounded-full text-sm font-semibold",children:I.length})]}),(0,t.jsxs)("div",{className:"space-y-2",children:[I.map((r,s)=>(0,t.jsxs)("button",{onClick:()=>y(r),className:`w-full text-left p-4 rounded-xl transition-all duration-300 group ${e===r?"bg-primary/5 border-2 border-primary shadow-md scale-[1.02]":"hover:bg-muted/50 border-2 border-transparent hover:border-border"}`,children:[(0,t.jsxs)("div",{className:"flex items-center justify-between mb-2",children:[(0,t.jsx)("span",{className:`px-3 py-1 text-xs font-bold rounded-lg ${A(r.method)} shadow-sm`,children:r.method}),r.auth&&(0,t.jsxs)("span",{className:"text-xs text-muted-foreground font-medium flex items-center gap-1",children:[(0,t.jsx)(u.LockClosedIcon,{className:"h-3 w-3"}),r.auth]})]}),(0,t.jsx)("div",{className:"text-sm font-mono text-foreground mb-2 break-all font-semibold",children:r.path}),(0,t.jsx)("div",{className:"text-xs text-muted-foreground line-clamp-2",children:r.description})]},s)),0===I.length&&(0,t.jsxs)("div",{className:"text-center py-12 text-muted-foreground",children:[(0,t.jsx)(c.MagnifyingGlassIcon,{className:"h-12 w-12 mx-auto mb-3 opacity-50"}),(0,t.jsx)("p",{className:"text-sm",children:"No endpoints found"})]})]})]})}),(0,t.jsx)("div",{className:"lg:col-span-2",children:e?(0,t.jsxs)("div",{className:"bg-card rounded-2xl shadow-xl border border-border p-8",children:[e.category&&(0,t.jsx)("div",{className:"mb-4",children:(0,t.jsxs)("span",{className:"inline-flex items-center gap-2 text-sm font-semibold text-primary bg-primary/10 px-4 py-2 rounded-lg border border-primary/20",children:[E(e.category),e.category]})}),(0,t.jsxs)("div",{className:"flex items-center gap-4 mb-6 pb-6 border-b-2 border-border",children:[(0,t.jsx)("span",{className:`px-4 py-2 text-sm font-bold rounded-lg ${A(e.method)} shadow-md`,children:e.method}),(0,t.jsx)("code",{className:"text-xl font-mono text-foreground break-all font-bold",children:e.path})]}),(0,t.jsx)("p",{className:"text-muted-foreground mb-8 text-lg leading-relaxed",children:e.description}),e.auth&&(0,t.jsxs)("div",{className:"mb-8 bg-amber-500/10 border-l-4 border-amber-500 rounded-r-xl p-6",children:[(0,t.jsxs)("h3",{className:"text-lg font-bold text-foreground mb-3 flex items-center gap-2",children:[(0,t.jsx)(u.LockClosedIcon,{className:"h-5 w-5 text-amber-500"}),"Authentication"]}),(0,t.jsx)("p",{className:"text-sm text-muted-foreground",children:"Required"===e.auth?"🔒 This endpoint requires authentication. Include the Bearer token in the Authorization header.":"🔓 Authentication is optional for this endpoint. Include the Bearer token for enhanced features."})]}),e.parameters&&e.parameters.length>0&&(0,t.jsxs)("div",{className:"mb-8",children:[(0,t.jsxs)("h3",{className:"text-lg font-bold text-foreground mb-4 flex items-center gap-2",children:[(0,t.jsx)(l,{className:"h-5 w-5 text-primary"}),"Parameters"]}),(0,t.jsx)("div",{className:"overflow-hidden rounded-xl border-2 border-border",children:(0,t.jsxs)("table",{className:"w-full text-sm",children:[(0,t.jsx)("thead",{children:(0,t.jsxs)("tr",{className:"bg-muted/50 border-b-2 border-border",children:[(0,t.jsx)("th",{className:"text-left py-4 px-4 font-bold text-muted-foreground",children:"Name"}),(0,t.jsx)("th",{className:"text-left py-4 px-4 font-bold text-muted-foreground",children:"Type"}),(0,t.jsx)("th",{className:"text-left py-4 px-4 font-bold text-muted-foreground",children:"Required"}),(0,t.jsx)("th",{className:"text-left py-4 px-4 font-bold text-muted-foreground",children:"Description"})]})}),(0,t.jsx)("tbody",{children:e.parameters.map((e,r)=>(0,t.jsxs)("tr",{className:`border-b border-border ${r%2==0?"bg-card":"bg-muted/20"} hover:bg-muted/50 transition-colors`,children:[(0,t.jsx)("td",{className:"py-4 px-4 font-mono font-semibold text-primary",children:e.name}),(0,t.jsx)("td",{className:"py-4 px-4 text-muted-foreground",children:e.type}),(0,t.jsx)("td",{className:"py-4 px-4",children:(0,t.jsx)("span",{className:`px-3 py-1 text-xs font-bold rounded-full ${e.required?"bg-destructive/10 text-destructive border border-destructive/20":"bg-muted text-muted-foreground border border-border"}`,children:e.required?"Yes":"No"})}),(0,t.jsx)("td",{className:"py-4 px-4 text-muted-foreground",children:e.description})]},r))})]})})]}),e.requestBody&&(0,t.jsxs)("div",{className:"mb-8",children:[(0,t.jsxs)("h3",{className:"text-lg font-bold text-foreground mb-4 flex items-center gap-2",children:[(0,t.jsx)(l,{className:"h-5 w-5 text-primary"}),"Request Body"]}),(0,t.jsxs)("p",{className:"text-sm text-muted-foreground mb-3 font-medium",children:["Content-Type: ",(0,t.jsx)("code",{className:"bg-muted px-2 py-1 rounded text-primary",children:e.requestBody.contentType})]}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono",children:"JSON"}),(0,t.jsx)("button",{onClick:()=>S(e.requestBody.schema,"request-schema"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"request-schema"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-8",children:(0,t.jsx)("code",{children:e.requestBody.schema})})]})]}),(0,t.jsxs)("div",{className:"mb-8",children:[(0,t.jsxs)("h3",{className:"text-lg font-bold text-foreground mb-4 flex items-center gap-2",children:[(0,t.jsx)(l,{className:"h-5 w-5 text-primary"}),"Responses"]}),(0,t.jsx)("div",{className:"space-y-4",children:e.responses.map((e,r)=>(0,t.jsxs)("div",{className:"border-2 border-border rounded-xl p-6 hover:border-primary/50 transition-all hover:shadow-lg bg-card",children:[(0,t.jsxs)("div",{className:"flex items-center gap-3 mb-4",children:[(0,t.jsx)("span",{className:`px-4 py-2 text-sm font-bold rounded-lg shadow-md ${e.status<300?"bg-green-500 text-white":e.status<400?"bg-blue-500 text-white":e.status<500?"bg-yellow-500 text-white":"bg-red-500 text-white"}`,children:e.status}),(0,t.jsx)("span",{className:"text-sm text-foreground font-medium",children:e.description})]}),e.schema&&(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-lg p-4 relative shadow-xl border border-gray-700",children:[(0,t.jsx)("button",{onClick:()=>S(e.schema,`response-${r}`),className:"absolute top-3 right-3 p-1.5 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:f===`response-${r}`?(0,t.jsx)(d.CheckIcon,{className:"h-4 w-4 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-4 w-4"})}),(0,t.jsx)("pre",{className:"text-green-400 text-xs overflow-x-auto font-mono leading-relaxed pr-8",children:(0,t.jsx)("code",{children:e.schema})})]})]},r))})]}),e.example&&(0,t.jsxs)("div",{className:"mb-6",children:[(0,t.jsxs)("h3",{className:"text-lg font-bold text-foreground mb-4 flex items-center gap-2",children:[(0,t.jsx)(g.SparklesIcon,{className:"h-5 w-5 text-primary"}),"Example"]}),e.example.request&&(0,t.jsxs)("div",{className:"mb-6",children:[(0,t.jsx)("h4",{className:"text-sm font-bold text-muted-foreground mb-3 flex items-center gap-2",children:"📤 Request:"}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono",children:"JSON"}),(0,t.jsx)("button",{onClick:()=>S(e.example.request,"example-request"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"example-request"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-8",children:(0,t.jsx)("code",{children:e.example.request})})]})]}),(0,t.jsxs)("div",{children:[(0,t.jsx)("h4",{className:"text-sm font-bold text-muted-foreground mb-3 flex items-center gap-2",children:"📥 Response:"}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono",children:"JSON"}),(0,t.jsx)("button",{onClick:()=>S(e.example.response,"example-response"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"example-response"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-8",children:(0,t.jsx)("code",{children:e.example.response})})]})]})]})]}):(0,t.jsx)("div",{className:"bg-card rounded-2xl shadow-xl border border-border p-16 text-center",children:(0,t.jsxs)("div",{className:"max-w-md mx-auto",children:[(0,t.jsx)("div",{className:"w-20 h-20 bg-primary/10 rounded-2xl flex items-center justify-center mx-auto mb-6",children:(0,t.jsx)(o.CodeBracketIcon,{className:"h-10 w-10 text-primary"})}),(0,t.jsx)("h3",{className:"text-2xl font-bold text-foreground mb-3",children:"Select an Endpoint"}),(0,t.jsx)("p",{className:"text-muted-foreground text-lg",children:"Choose an endpoint from the list to view detailed documentation and examples."})]})})})]}),(0,t.jsxs)("div",{className:"mt-16 bg-muted/30 rounded-3xl shadow-2xl border-2 border-border p-10 relative overflow-hidden",children:[(0,t.jsx)("div",{className:"absolute top-0 right-0 w-64 h-64 bg-primary/5 rounded-full blur-3xl"}),(0,t.jsx)("div",{className:"absolute bottom-0 left-0 w-64 h-64 bg-primary/5 rounded-full blur-3xl"}),(0,t.jsxs)("div",{className:"relative z-10",children:[(0,t.jsxs)("div",{className:"flex items-center gap-4 mb-6",children:[(0,t.jsx)("div",{className:"p-4 bg-primary rounded-2xl shadow-lg",children:(0,t.jsx)(o.CodeBracketIcon,{className:"h-10 w-10 text-primary-foreground"})}),(0,t.jsxs)("div",{children:[(0,t.jsx)("h2",{className:"text-4xl font-bold text-foreground",children:"Integration Guide"}),(0,t.jsx)("p",{className:"text-muted-foreground text-lg mt-1",children:"Learn how to integrate Trunc into your applications"})]})]}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-amber-50 to-orange-50 border-2 border-amber-300 rounded-2xl p-8 mb-8 shadow-lg",children:[(0,t.jsxs)("h3",{className:"text-2xl font-bold text-gray-900 mb-5 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-3xl",children:"⚠️"})," Prerequisites"]}),(0,t.jsxs)("p",{className:"text-sm text-gray-700 mb-6 leading-relaxed",children:["Before you can use the API, you need a Trunc account and an API key. ",(0,t.jsx)("strong",{children:"You cannot register or login via the API."})," Account creation and API key generation must be done through the web interface."]}),(0,t.jsxs)("div",{className:"space-y-5",children:[(0,t.jsxs)("div",{className:"flex gap-4 items-start group hover:translate-x-1 transition-transform",children:[(0,t.jsx)("div",{className:"flex-shrink-0 w-10 h-10 bg-gradient-to-br from-primary-600 to-primary-700 text-white rounded-xl flex items-center justify-center font-bold text-lg shadow-lg",children:"1"}),(0,t.jsxs)("div",{children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 text-lg mb-1",children:"Register on the Website"}),(0,t.jsxs)("p",{className:"text-sm text-gray-700",children:["Visit ",(0,t.jsx)("a",{href:"/",className:"text-primary-600 hover:text-primary-700 underline font-semibold",children:"trunc.site"})," and create an account if you haven't already."]})]})]}),(0,t.jsxs)("div",{className:"flex gap-4 items-start group hover:translate-x-1 transition-transform",children:[(0,t.jsx)("div",{className:"flex-shrink-0 w-10 h-10 bg-gradient-to-br from-primary-600 to-primary-700 text-white rounded-xl flex items-center justify-center font-bold text-lg shadow-lg",children:"2"}),(0,t.jsxs)("div",{children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 text-lg mb-1",children:"Login to Dashboard"}),(0,t.jsx)("p",{className:"text-sm text-gray-700",children:"Sign in to your account and access the dashboard."})]})]}),(0,t.jsxs)("div",{className:"flex gap-4 items-start group hover:translate-x-1 transition-transform",children:[(0,t.jsx)("div",{className:"flex-shrink-0 w-10 h-10 bg-gradient-to-br from-primary-600 to-primary-700 text-white rounded-xl flex items-center justify-center font-bold text-lg shadow-lg",children:"3"}),(0,t.jsxs)("div",{children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 text-lg mb-1",children:"Generate API Key"}),(0,t.jsx)("p",{className:"text-sm text-gray-700",children:"Navigate to Settings → API Keys and create a new API key with the required permissions."})]})]})]})]}),(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 mb-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h3",{className:"text-2xl font-bold text-gray-900 mb-5 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-3xl",children:"🚀"})," Quick Start"]}),(0,t.jsx)("p",{className:"text-sm text-gray-600 mb-6 leading-relaxed",children:"Once you have your API key from the prerequisites above, follow these steps:"}),(0,t.jsxs)("div",{className:"space-y-5",children:[(0,t.jsxs)("div",{className:"flex gap-5 items-start group",children:[(0,t.jsx)("div",{className:"flex-shrink-0 w-12 h-12 bg-gradient-to-br from-green-500 to-emerald-600 text-white rounded-full flex items-center justify-center font-bold text-lg shadow-lg",children:"1"}),(0,t.jsxs)("div",{className:"flex-1",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"Copy Your API Key"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"After creating an API key in the dashboard, copy it immediately. You won't be able to see it again!"})]})]}),(0,t.jsxs)("div",{className:"flex gap-5 items-start group",children:[(0,t.jsx)("div",{className:"flex-shrink-0 w-12 h-12 bg-gradient-to-br from-blue-500 to-indigo-600 text-white rounded-full flex items-center justify-center font-bold text-lg shadow-lg",children:"2"}),(0,t.jsxs)("div",{className:"flex-1",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"Store Securely"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"Save your API key in a secure location (environment variables, secrets manager, etc.). Never commit it to version control!"})]})]}),(0,t.jsxs)("div",{className:"flex gap-5 items-start group",children:[(0,t.jsx)("div",{className:"flex-shrink-0 w-12 h-12 bg-gradient-to-br from-purple-500 to-pink-600 text-white rounded-full flex items-center justify-center font-bold text-lg shadow-lg",children:"3"}),(0,t.jsxs)("div",{className:"flex-1",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"Make Your First Request"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"Use the code examples below to start shortening URLs programmatically!"})]})]})]})]}),(0,t.jsxs)("div",{className:"mb-8",children:[(0,t.jsxs)("div",{className:"flex items-center gap-3 mb-6",children:[(0,t.jsx)("h3",{className:"text-2xl font-bold text-gray-900",children:"Code Examples"}),(0,t.jsx)("div",{className:"flex-1"}),(0,t.jsx)("div",{className:"flex gap-2",children:["javascript","react","nodejs","python"].map(e=>(0,t.jsx)("button",{onClick:()=>C(e),className:`px-5 py-2.5 rounded-xl font-semibold text-sm transition-all duration-300 ${L===e?"bg-gradient-to-r from-primary-600 to-blue-600 text-white shadow-lg scale-105":"bg-white text-gray-700 border-2 border-gray-200 hover:border-primary-300 hover:text-primary-700"}`,children:"javascript"===e?"JavaScript":"react"===e?"React":"nodejs"===e?"Node.js":"Python"},e))})]}),"javascript"===L&&(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h4",{className:"text-xl font-bold text-gray-900 mb-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-2xl",children:"📜"})," Vanilla JavaScript"]}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-3",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg",children:"JavaScript"}),(0,t.jsx)("button",{onClick:()=>S(`async function shortenURL(originalUrl) {
  const response = await fetch('${U}/api/v1/shorten', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer YOUR_API_KEY_HERE'
    },
    body: JSON.stringify({ url: originalUrl })
  });

  if (!response.ok) {
    throw new Error('Failed to shorten URL');
  }

  const data = await response.json();
  return data.short_url;
}

// Usage
shortenURL('https://example.com/very-long-url')
  .then(shortUrl => console.log('Shortened URL:', shortUrl))
  .catch(error => console.error('Error:', error));`,"js-example"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"js-example"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10",children:(0,t.jsx)("code",{children:`async function shortenURL(originalUrl) {
  const response = await fetch('${U}/api/v1/shorten', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': 'Bearer YOUR_API_KEY_HERE'
    },
    body: JSON.stringify({ url: originalUrl })
  });

  if (!response.ok) {
    throw new Error('Failed to shorten URL');
  }

  const data = await response.json();
  return data.short_url;
}

// Usage
shortenURL('https://example.com/very-long-url')
  .then(shortUrl => console.log('Shortened URL:', shortUrl))
  .catch(error => console.error('Error:', error));`})})]})]}),"react"===L&&(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h4",{className:"text-xl font-bold text-gray-900 mb-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-2xl",children:"⚛️"})," React Custom Hook"]}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-3",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg",children:"React"}),(0,t.jsx)("button",{onClick:()=>S(`import { useState } from 'react';

export function useURLShortener(apiKey) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const shortenURL = async (originalUrl) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('${U}/api/v1/shorten', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        },
        body: JSON.stringify({ url: originalUrl })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to shorten URL');
      }

      const data = await response.json();
      return data.short_url;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { shortenURL, loading, error };
}`,"react-hook"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"react-hook"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10",children:(0,t.jsx)("code",{children:`import { useState } from 'react';

export function useURLShortener(apiKey) {
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState(null);

  const shortenURL = async (originalUrl) => {
    setLoading(true);
    setError(null);

    try {
      const response = await fetch('${U}/api/v1/shorten', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        },
        body: JSON.stringify({ url: originalUrl })
      });

      if (!response.ok) {
        const errorData = await response.json();
        throw new Error(errorData.message || 'Failed to shorten URL');
      }

      const data = await response.json();
      return data.short_url;
    } catch (err) {
      setError(err.message);
      throw err;
    } finally {
      setLoading(false);
    }
  };

  return { shortenURL, loading, error };
}`})})]})]}),"nodejs"===L&&(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h4",{className:"text-xl font-bold text-gray-900 mb-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-2xl",children:"🟢"})," Node.js Backend"]}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-3",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg",children:"Node.js"}),(0,t.jsx)("button",{onClick:()=>S(`const axios = require('axios');

async function shortenURL(originalUrl, apiKey) {
  try {
    const response = await axios.post('${U}/api/v1/shorten',
      { url: originalUrl },
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        }
      }
    );

    return response.data.short_url;
  } catch (error) {
    console.error('Error shortening URL:', error.response?.data || error.message);
    throw error;
  }
}

// Express.js endpoint example:
app.post('/api/shorten', async (req, res) => {
  const { url } = req.body;
  const apiKey = process.env.TRUNC_API_KEY;

  try {
    const shortUrl = await shortenURL(url, apiKey);
    res.json({ success: true, shortUrl });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: 'Failed to shorten URL'
    });
  }
});`,"node-example"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"node-example"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10",children:(0,t.jsx)("code",{children:`const axios = require('axios');

async function shortenURL(originalUrl, apiKey) {
  try {
    const response = await axios.post('${U}/api/v1/shorten',
      { url: originalUrl },
      {
        headers: {
          'Content-Type': 'application/json',
          'Authorization': \`Bearer \${apiKey}\`
        }
      }
    );

    return response.data.short_url;
  } catch (error) {
    console.error('Error shortening URL:', error.response?.data || error.message);
    throw error;
  }
}

// Express.js endpoint example:
app.post('/api/shorten', async (req, res) => {
  const { url } = req.body;
  const apiKey = process.env.TRUNC_API_KEY;

  try {
    const shortUrl = await shortenURL(url, apiKey);
    res.json({ success: true, shortUrl });
  } catch (error) {
    res.status(500).json({
      success: false,
      error: 'Failed to shorten URL'
    });
  }
});`})})]})]}),"python"===L&&(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h4",{className:"text-xl font-bold text-gray-900 mb-4 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-2xl",children:"🐍"})," Python Integration"]}),(0,t.jsxs)("div",{className:"bg-gradient-to-br from-gray-900 to-gray-800 rounded-xl p-6 relative shadow-2xl border border-gray-700",children:[(0,t.jsxs)("div",{className:"absolute top-4 right-4 flex items-center gap-3",children:[(0,t.jsx)("span",{className:"text-xs text-gray-400 font-mono bg-gray-800 px-3 py-1 rounded-lg",children:"Python"}),(0,t.jsx)("button",{onClick:()=>S(`import requests
import os

def shorten_url(original_url, api_key):
    """Shorten a URL using the Trunc API"""
    response = requests.post(
        '${U}/api/v1/shorten',
        json={'url': original_url},
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}'
        }
    )

    if response.status_code == 201:
        return response.json()['short_url']
    else:
        raise Exception(f"Failed to shorten URL: {response.text}")

# Usage
api_key = os.getenv('TRUNC_API_KEY')
short_url = shorten_url('https://example.com/long-url', api_key)
print(f"Shortened URL: {short_url}")`,"python-example"),className:"p-2 text-gray-400 hover:text-white hover:bg-white/10 rounded-lg transition-all",children:"python-example"===f?(0,t.jsx)(d.CheckIcon,{className:"h-5 w-5 text-green-400"}):(0,t.jsx)(i.ClipboardDocumentIcon,{className:"h-5 w-5"})})]}),(0,t.jsx)("pre",{className:"text-green-400 text-sm overflow-x-auto font-mono leading-relaxed pt-10",children:(0,t.jsx)("code",{children:`import requests
import os

def shorten_url(original_url, api_key):
    """Shorten a URL using the Trunc API"""
    response = requests.post(
        '${U}/api/v1/shorten',
        json={'url': original_url},
        headers={
            'Content-Type': 'application/json',
            'Authorization': f'Bearer {api_key}'
        }
    )

    if response.status_code == 201:
        return response.json()['short_url']
    else:
        raise Exception(f"Failed to shorten URL: {response.text}")

# Usage
api_key = os.getenv('TRUNC_API_KEY')
short_url = shorten_url('https://example.com/long-url', api_key)
print(f"Shortened URL: {short_url}")`})})]})]})]}),(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 mb-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h3",{className:"text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-3xl",children:"🔐"})," Security Best Practices"]}),(0,t.jsxs)("div",{className:"space-y-5",children:[(0,t.jsxs)("div",{className:"border-l-4 border-red-500 bg-red-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg flex items-center gap-2",children:"⛔ Never Expose API Keys in Frontend Code"}),(0,t.jsx)("p",{className:"text-sm text-gray-700",children:"API keys should NEVER be included directly in your frontend JavaScript code or committed to version control."})]}),(0,t.jsxs)("div",{className:"border-l-4 border-green-500 bg-green-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-3 text-lg flex items-center gap-2",children:"✅ Use Environment Variables"}),(0,t.jsx)("p",{className:"text-sm text-gray-700 mb-3",children:"Store API keys in environment variables:"}),(0,t.jsx)("code",{className:"block bg-gray-900 text-green-400 p-4 rounded-lg text-sm font-mono",children:"TRUNC_API_KEY=us_your_api_key_here"})]}),(0,t.jsxs)("div",{className:"border-l-4 border-blue-500 bg-blue-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg flex items-center gap-2",children:"✅ Implement a Backend Proxy"}),(0,t.jsx)("p",{className:"text-sm text-gray-700",children:"Create a backend endpoint that calls the Trunc API. Your frontend calls your backend, which securely stores the API key."})]}),(0,t.jsxs)("div",{className:"border-l-4 border-purple-500 bg-purple-50 rounded-r-xl p-6 hover:shadow-lg transition-shadow",children:[(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg flex items-center gap-2",children:"✅ Use Permission Scopes"}),(0,t.jsx)("p",{className:"text-sm text-gray-700",children:"Create API keys with only the permissions they need. Don't give full access if you only need URL shortening."})]})]})]}),(0,t.jsxs)("div",{className:"bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100",children:[(0,t.jsxs)("h3",{className:"text-2xl font-bold text-gray-900 mb-6 flex items-center gap-2",children:[(0,t.jsx)("span",{className:"text-3xl",children:"💡"})," Common Use Cases"]}),(0,t.jsxs)("div",{className:"grid md:grid-cols-2 gap-6",children:[(0,t.jsxs)("div",{className:"border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group",children:[(0,t.jsx)("div",{className:"text-4xl mb-3 group-hover:scale-110 transition-transform",children:"📝"}),(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"Blog Post Sharing"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"Automatically generate short URLs for blog posts to share on social media."})]}),(0,t.jsxs)("div",{className:"border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group",children:[(0,t.jsx)("div",{className:"text-4xl mb-3 group-hover:scale-110 transition-transform",children:"📧"}),(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"Email Newsletters"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"Create trackable short links for email campaigns with custom UTM parameters."})]}),(0,t.jsxs)("div",{className:"border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group",children:[(0,t.jsx)("div",{className:"text-4xl mb-3 group-hover:scale-110 transition-transform",children:"🎯"}),(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"Marketing Campaigns"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"Generate branded short URLs for marketing materials and track click performance."})]}),(0,t.jsxs)("div",{className:"border-2 border-gray-200 rounded-xl p-6 hover:border-primary-400 hover:shadow-xl transition-all group",children:[(0,t.jsx)("div",{className:"text-4xl mb-3 group-hover:scale-110 transition-transform",children:"📱"}),(0,t.jsx)("h4",{className:"font-bold text-gray-900 mb-2 text-lg",children:"QR Code Generation"}),(0,t.jsx)("p",{className:"text-sm text-gray-600",children:"Create short URLs that are perfect for encoding in QR codes."})]})]})]})]})]})]})]})})}e.s(["default",()=>y],36791)}]);