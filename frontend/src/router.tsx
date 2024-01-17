import { Outlet, Router, Route, RootRoute } from "@tanstack/react-router";
import Layout from "./components/Layout/Layout.tsx";
import Index from "./pages/Index.tsx";
import QuestionForm from "./pages/Registration.tsx";
import { UploadDesign } from "./pages/UploadDesign.tsx";
import { Login } from "./pages/Login.tsx";
import Jobs from "./pages/Jobs.tsx";
import JobInfo from "./pages/JobAccept.tsx";
import ProfilePage from "./pages/Profile.tsx";
import JobView from "./pages/JobView.tsx";

const rootRoute = new RootRoute({
  component: () => (
    <>
      <Layout>
        <Outlet />
      </Layout>
      {
        import.meta.env.DEV
        // && <TanStackRouterDevtools />
      }
    </>
  ),
});

// Create new routes here
const error404Route = new Route({
  getParentRoute: () => rootRoute,
  path: "*",
  component: () => <h1>404</h1>, //TODO: Make it pretty! (Also define as a component in pages dir)
});

const indexRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/",
  component: Index,
});

const loginRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/login",
  component: Login,
});

const profileRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/profile",
  component: ProfilePage,
});

const jobsRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/jobs",
  component: Jobs,
});

const jobAcceptRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/job-accept/$jobId",
  component: JobInfo,
});

const jobViewRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/jobs/$jobId",
  component: JobView,
});

const registrationRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/register",
  component: QuestionForm,
});

const uploadDesignRoute = new Route({
  getParentRoute: () => rootRoute,
  path: "/upload-design",
  component: UploadDesign,
});

// Add routes to the route tree here
const routeTree = rootRoute.addChildren([
  error404Route,
  indexRoute,
  loginRoute,
  profileRoute,
  registrationRoute,
  uploadDesignRoute,
  jobsRoute,
  jobAcceptRoute,
  jobViewRoute,
]);

const router = new Router({ routeTree });
declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}
export default router;
